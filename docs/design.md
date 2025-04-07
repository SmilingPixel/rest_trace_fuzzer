# [WIP] 技术方案 trace_rest_test


# 1. 背景

Repository: [GitHub - SmilingPixel/rest_trace_fuzzer](https://github.com/SmilingPixel/rest_trace_fuzzer)


# 2. 总体架构

![You can download the raw file, and edit using draw.io](architecture.svg)

The architecture diagram is powered by [draw.io](https://app.diagrams.net/).

总体分为3个模块

## 2.1. 预处理和解析

| Module      | Description                                      | Note                                                                 |
|-------------|--------------------------------------------------|----------------------------------------------------------------------|
| OpenAPI Parser | 解析 OpenAPI 文件，生成接口定义                      | 输入：一批 OpenAPI 定义，包括 1. 每个 service 的接口定义（包括 HTTP 和 gRPC） 2. 整个系统对外暴露的接口定义 输出：解析完成的接口，数据结构设计见后面章节 |
| ODG Parser  | 解析系统外部暴露 OpenAPI，生成 ODG (Operation Dependency Graph) | 输入：解析完成的 OpenAPI 输出：ODG，包括 1. 整个系统对外暴露的接口之间的依赖关系 2. 每个 service 内部的接口之间的依赖关系（即对于图中的任何一条边，其两端的节点必须属于不同的 service） |
| DFG Parser  | 内部服务 API，生成 DFG (Data Flow Graph)            | 输入：解析完成的 OpenAPI 输出：DFG，包括 1. 每个 service 内部的接口之间的数据流传递关系（即对于图中的任何一条边，其两端的节点必须属于不同的 service） |

## 2.2. 数据持久化模块

| Module         | Description                                      | Note                                                                 |
|----------------|--------------------------------------------------|----------------------------------------------------------------------|
| Interfaces DB  | 存储解析完成的接口定义                              | 数据结构：见后面章节                                                   |
| ODG DB         | 存储 ODG                                           | 数据结构：见后面章节                                                   |
| DFG DB         | 存储 DFG                                           | 数据结构：见后面章节                                                   |
| Runtime Info Graph         | 存储程序运行时的 DFG 调用和覆盖状态                                          | 数据结构：见后面章节                                                   |
| Resource Pool  | 存储资源池，目前主要用于存储测试过程中创建的资源       |          |
| Testcase Queue     | 存储测试用例，优先级队列，根据选取策略的不同，依据各种指标（例如覆盖率提升等）计算优先级 | |
| Operation Case Queues     | 存储候选请求，多个优先级队列，每种请求（Endpoint）一个队列，根据选取策略的不同，依据各种指标（例如覆盖率提升等）计算优先级 | |
| Trace Manager   | 拉取、解析和处理 trace | |


## 2.3. 测试执行模块

| Module               | Description                                      | Note                                                                 |
|----------------------|--------------------------------------------------|----------------------------------------------------------------------|
| Testcase Scheduler  | 根据 ODG, DFG, Runtime Info Graph 等，调度测试用例的优先级                        | 输入：数据持久化模块中的数据 输出：本次预计执行的接口序列         |
| Operation Population | 根据接口定义，填充参数，实例化接口                           | 输入：接口定义，资源池数据 输出：接口请求实例                           |
| Test Driver          | 执行接口请求，记录执行结果，处理反馈，更新数据持久化模块中的数据 |                                                                      |
| Response Processer       | 收集系统响应，存储和初步检查测试结果        |                                                                      |
| Trace Analyser       | 收集 Trace，同步更新 RuntimeInfo Graph 的状态        |                                                                      |


## 2.4. 策略执行模块

定义各类操作的策略，将策略与执行动作分离，便于后期进一步扩展
| Module               | Description                                      | Note                                                                 |
|----------------------|--------------------------------------------------|----------------------------------------------------------------------|
| Strategist           | 提供对外接口，定义和管理测试策略                   | 所有的策略管理和使用都必须经过其接口，收敛权责                                      |
| Value Generation     | 负责生成测试值，支持多种值生成策略                 | 例如随机生成、从资源池获取、是否需要 mutation 等                            |
| Resource Mutation     | 负责测试值的 Mutation，支持多种 Mutation 策略                 | 例如值的随机改变、请求体结构的随机变换 等                            |
| TODO     | TODO                 | @xunzhou24                            |


## 2.5. 结果报告模块

| Module               | Description                                      | Note                                                                 |
|----------------------|--------------------------------------------------|----------------------------------------------------------------------|
| Internal Service Reporter | 输出内部接口测试结果，包括覆盖率、错误率等指标 | 输入：测试执行模块中的数据 输出：内部接口测试报告 |
| System Reporter      | 输出外部接口测试结果，包括覆盖率、错误率等指标 | 输入：测试执行模块中的数据 输出：外部接口测试报告 |
| Fuzzer State Reporter | 输出 Fuzzer 状态，包括资源池状况等 | 目前主要用于记录执行过程中的详细信息，便于开发和调试 |
| Test Log Reporter    | 输出测试日志，包括测试用例执行情况、错误信息等 | 需要嵌入在主 fuzzer 中，执行过程中实时记录产生的原始测试用例和执行情况 |


# 3. 具体实现方案

下面是项目的代码结构：
```bash
.
├── cmd
│   └── api-fuzzer          # Main entry point for the API fuzzer application
│       └── main.go         # Main application logic
├── config
│   └── config.json         # Configuration file for the application
├── docs                    # Documentation files
│   ├── architecture.drawio # Architecture diagram source file
│   ├── architecture.svg    # Architecture diagram in SVG format
│   └── design.md           # Design documentation
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies file
├── internal                # Internal packages
│   ├── config              # Configuration related code
│   │   ├── arg_config_generate.py # Script to generate argument configuration
│   │   ├── arg_config.json  # Argument configuration file, auto-generated
│   │   ├── arg_parse.go     # Argument parsing logic, auto-generated
│   │   └── config.go
│   └── fuzzer              # Fuzzer related code
│       ├── basic_fuzzer.go  # Basic fuzzer implementation
│       ├── config.go        # Fuzzer configuration handling
│       ├── fuzzer.go        # Main fuzzer interfaces
│       └── snapshot.go      # Snapshot handling for fuzzing
├── LICENSE                 # License file
├── Makefile                # Makefile for generate and run automation
├── pkg                     # Public packages
│   ├── casemanager         # Case management logic
│   │   ├── case.go          # Case structure and methods
│   │   └── case_manager.go  # Case manager implementation
│   ├── feedback            # Feedback handling logic
│   │   ├── response_process.go # Response processing logic
│   │   ├── runtime_graph.go  # Runtime graph handling
│   │   └── trace           # Trace handling logic
│   │       ├── model.go      # Trace model definitions
│   │       ├── trace_db.go   # Trace database interactions
│   │       ├── trace_fetcher.go # Trace fetching logic
│   │       └── trace_manager.go # Trace manager implementation
│   ├── parser              # Parsing logic
│   │   ├── api_parser.go    # API parsing logic
│   │   ├── dependency_parser.go # Dependency parsing logic
│   │   └── dependency_restler_parser.go # RESTler dependency parsing logic
│   ├── report              # Reporting logic
│   │   ├── fuzzer_state_reporter.go # Fuzzer state reporting
│   │   ├── internal_service_reporter.go # Internal service reporting
│   │   ├── models.go        # Report models
│   │   ├── system_reporter.go # System reporting logic
│   │   └── test_log_reporter.go # Test log reporting
│   ├── resource            # Resource management logic
│   │   ├── resource.go      # Resource structure and methods
│   │   └── resource_manager.go # Resource manager implementation
│   ├── static              # Static info related
│   │   ├── api_manager.go   # API manager logic
│   │   ├── dependency.go    # Dependency handling
│   │   ├── dfg.go           # Data flow graph handling
│   │   └── simple_model.go  # Simple model for static info
│   ├── strategy             # Fuzzing strategies
│   │   ├── fuzz_strategist.go  # Fuzzing strategist implementation
│   │   ├── resource_mutate.go  # strategies for resource mutation
│   │   └── value_generate.go   # strategies for value generation
│   └── utils               # Utility functions
│       ├── http            # HTTP-related utilities
│       │   ├── http_utils.go  # HTTP utility functions
│       │   └── middleware.go  # Middleware functions using in http utils
│       ├── common_utils.go  # Common utility functions
│       ├── nlp_utils.go     # NLP utility functions
│       └── openapi_utils.go # OpenAPI utility functions
├── README.md               # Project README file
└── scripts                 # Scripts for various tasks
  ├── build.sh              # Build script
  ├── clean_build.sh        # Script to clean build artifacts
  ├── clean_output.sh       # Script to clean output
  ├── generate_arg_config_code.sh # Script to generate argument configuration code
  ├── include               # Protobuf include files
  │   └── google
  │       └── api
  │           ├── annotations.proto # Protobuf annotations for API
  │           └── http.proto        # Protobuf definitions for HTTP
  ├── internal_service_report_visualize.py # Script to visualize internal service reports
  ├── proto_gen_oas.sh      # Script to generate OpenAPI specs from protobuf
  └── run.sh                # Script to run the application
```
该工具的主要部分在下面两个文件夹下:
- `/pkg`: 主要模块设计实现
- `/internal`: 工具内部流程相关

## 3.1. OpenAPI 解析

框架选取 [getkin/kin-openapi](https://github.com/getkin/kin-openapi/)

解析时，对于输入的 OpenAPI Spec，构建两套数据：
1. 整个系统的接口定义
2. 每个 service 的接口定义，存储为一个 map，key 为 service name


## 3.2. ODG, DFG 解析

### 3.2.1. 数据结构

#### 3.2.1.1. ODG

Operation 定义粒度: (path, method)，例如`(/api/v1/pet/{id}, GET)`

Edge， a -> b 表示 a 依赖于 b，即 a 的执行需要 b 的执行结果
| Field            | Type       | Description                         |
|------------------|------------|-------------------------------------|
| source           | Operation  | 表示依赖的源 Operation               |
| target           | Operation  | 表示依赖的目标 Operation             |
| source_resource  | [TODO: 待定] | 表示源 Operation 依赖的资源（例如某个参数） |

Graph，包含一组 Edge，表示整个系统外部API之间的依赖关系。

#### 3.2.1.2. DFG

Operation 定义粒度: (service, path, method)，例如`(pet_service, /api/v1/pet/{id}, GET)`

Edge， a -> b 表示存在数据流从 a 到 b
| Field            | Type       | Description                         |
|------------------|------------|-------------------------------------|
| source           | Operation  | 表示源 Operation               |
| target           | Operation  | 表示目标 Operation             |
| source_property  | APIProperty | 表示存在流动的（源）资源（例如某个参数） |
| target_property  | APIProperty | 表示存在流动的（目标）资源（例如某个参数） |

Graph，包含一组 Edge，表示整个系统内部API之间的数据流传递关系。

我们期望 source_property 和 target_property 是一样的，因为我们认为这两个资源是同一个资源，只是在不同的操作中传递。

Note: 由于对于一组 Operation，他们的一对 property 都是来自 request 或 response 的，因此总会有双向的边，即 a -> b 和 b -> a 都会存在。这部分需要更好的处理。@xunzhou24


### 3.2.2. ODG 解析

#### 3.2.2.1. 参考
参考 [SeUniVr/RestTestGen](https://github.com/SeUniVr/RestTestGen)，原论文:
```bibtex
@article{corradini2022nominalerror,
    doi = {10.1002/stvr.1808},
    url = {https://doi.org/10.1002/stvr.1808},
    year = {2022},
    month = jan,
    publisher = {Wiley},
    author = {Davide Corradini and Amedeo Zampieri and Michele Pasqua and Emanuele Viglianisi and Michael Dallago and Mariano Ceccato},
    title = {Automated black-box testing of nominal and error scenarios in RESTful APIs},
    journal = {Software Testing, Verification and Reliability}
}
```

核心方法 [extractDataDependencies](https://github.com/SeUniVr/RestTestGen/blob/363eebc9d8c26cb20a724e5dd58de1fb0cd1f346/src/main/java/io/resttestgen/core/operationdependencygraph/OperationDependencyGraph.java#L76)

```java
 // If input and output parameters have the same normalized name
if (outputParameter.getNormalizedName().equals(inputParameter.getNormalizedName())) {

    DependencyEdge edge = new DependencyEdge(outputParameter, inputParameter);

    graph.addEdge(getNodeFromOperation(targetOperation), getNodeFromOperation(sourceOperation), edge);
    commonParametersNames.add(outputParameter.getNormalizedName());
}
```

目前的实现，考虑到已经有大量成熟的工作了，因此目前支持直接导入 Restler 的解析文件，构建ODG。

### 3.2.3. DFG 解析

- 类似于 ODG 解析，目前设计主要依赖于资源名称的解析
  - 名称的匹配依赖字符串相似度，目前使用编辑距离（Levenshtein distance），未来预期加入多种启发式规则，可参考 Restler 的实现
- 更加侧重于“**数据流向**”的依赖关系，因此解析的资源对，要么同时来自于两个 Operation 的输入参数，要么同时来自于两个 Operation 的输出参数


#### 3.2.4. 实现差异

对于DFG，由于是内部 API 调用的依赖，因此更加侧重于“**资源流向**”的依赖关系，例如：
- CheckoutService 的 placeOrder 调用了 PaymentService 的 charge
- placeOrder 输入参数和 charge 输出参数中，都有 currency 字段
- 我们认为，这一信息，表示该数据从 CheckoutService 流向 PaymentService
- 因此，我们在 ODG 中，应该有一条边，表示 placeOrder 依赖于 charge

主要依赖于两条规则：
1. 如果两个参数的 normalized name 匹配（依据预定的各种规则），那么认为这两个参数有依赖关系
2. 两个操作之间存在“**资源流向**”的依赖关系

## 3.3. Resource Pool

### 3.3.1. 资源池的设计

一个 resource 为一个类 JSON 结构，可从 json 格式无损转换，即包含 object, array 和几种 primitive type 的类型。

Resource Pool 提供两种类型的查询:
1. 基于资源名的查询
2. 基于资源类型的查询（仅支持 primitive type）
例如:
```go
rsc1 := resourceManager.GetSingleResourceByType(SimpleAPIPropertyTypeString)
rsc2 := resourceManager.GetSingleResourceByName("Capitano")
```

### 3.3.2. 资源的生成

- 外部输入的字典，作为资源池的初始资源的一部分
- OpenAPI 中的 example 字段，作为资源池的初始资源的一部分(TODO: @xunzhou24)
- 测试过程中的 response，包含了请求的资源


## 3.4. Testcase Queue

总体思路：
- ~~从 ODG 中找到入度为 0 的节点，作为种子节点~~
- 从种子节点开始，根据 ODG 的依赖关系，考虑优先级，构建测试序列
- 将测试序列中的节点加入 Testcase Queue （类似于 DFS 的思路）
- 优先级（可以参考的方案）:
  - 提高内部 ODG 覆盖率的测试用例具有更高的优先级
  - 提高系统外部接口覆盖率（Path, Status Code）的测试用例具有更高的优先级
  - ...

### 3.4.1. 优先级的设计

- 测试用例中，序列本身和序列中的每个操作均有独立的优先级可供计算
- 参见[`case_manager.go`](../pkg/casemanager/case_manager.go)，目前优先级的计算主要依赖于*是否产生新的覆盖*（无论是内部服务的覆盖还是系统API的响应状态的覆盖），我们将依据这一变化，适当随机增减优先级
- 每轮测试执行完毕后，会根据优先级，对队列中的测试用例重新排序。

建议参考的方案：
- AFL
  - https://github.com/google/AFL/blob/master/afl-fuzz.c#L8091
  - https://zhuanlan.zhihu.com/p/624286070
  - https://www.zhihu.com/question/388240608/answer/1157919593
- LibFuzzer
  - https://github.com/llvm/llvm-project/blob/main/compiler-rt/lib/fuzzer/FuzzerLoop.cpp#L723

## 3.5. Operation Case Queues

由多个优先级队列组成，存储候选的操作（主要来自历史执行的操作），用于扩展测试序列时提供备选项。

数据结构上，是一个 map，key 为 Endpoint，value 为 list[Operation]。

例如，
```json
{
  "GET /api/pet": [...],
  "POST /api/user": [...]
}
```

## 3.6. Test Driver

测试执行，发送请求和处理响应，更新数据持久化模块中的数据


## 3.7. Trace Manager

拉取和解析 trace，并根据 trace 数据，生成预处理反馈，引导测试

### 3.7.1. Trace Fetcher

- 需要提前修改服务，使得我们能够从请求的响应（头或者响应体）中获取 trace ID
- OpenTelemetry Demo 中，使用 Jaeger 提供的 API 进行 trace 搜集
- 参考这个 [官方文档](https://www.jaegertracing.io/docs/2.4/apis/)

**Note**: 由于服务内部的 trace 收集等需要一定的时间，因此请求完后，我们并不能马上获取到 trace 数据，需要等待几秒钟，本项目中按照3秒钟的时间进行等待。


### 3.7.2. Trace 分析

1. 记录覆盖的调用边和调用次数，更新 DFG 实例
2. 统计覆盖率更新，上报给 Runtime Info Graph
3. 更新 DFG，补充缺失的依赖关系，修改错误的依赖关系 [TODO: 具体实现待定 @xunzhou24]


### 3.8. Runtime Info Graph

- 记录程序运行时的 DFG 调用和覆盖状态
- 基于 DFG，目前用于记录 DFG 上的调用次数，用于后续的覆盖率统计和测试用例优先级计算

Edge， a -> b 表示存在数据流从 a 到 b，同时携带了调用次数信息
| Field            | Type       | Description                         |
|------------------|------------|-------------------------------------|
| source           | Operation  | 表示源 Operation               |
| target           | Operation  | 表示目标 Operation             |
| hit_count       | int        | 表示调用次数                       |


## 3.8. 网络相关

### 3.8.1. HTTP 请求的实时修改

目前支持脚本语言 [Starlark](https://github.com/bazelbuild/starlark) 的脚本化支持，用于修改请求的参数，例如鉴权等。

具体实现见 [http client 中间件](../pkg/utils/http/middleware.go)。


### 3.8.2 HTTPS 支持

目前通过跳过证书验证的方式，支持 HTTPS 请求。（你就说有没有支持吧）


# 4. 子任务排期

- [x] 初步的技术方案设计，包括整体架构和模块设计
- [x] OpenAPI 解析模块的实现
- [x] RestTestGen 的 ODG 解析模块的移植
- [x] DFG 解析模块的实现
- [x] trace 拉取和解析模块的实现
- [x] 基础的 testcase queue 的实现
- [x] 资源池的基础实现
- [x] 基础的 fuzzing 循环启动
- [x] 完善的脚本化支持测试请求的修改，主要用于鉴权等
- [x] trace 反馈信息的处理
- [x] testcase queue 的优先级的基础实现
- [x] operation case queues 的设计与实现
- [x] 基础版本的 mutation 策略
- [ ] 内部服务接口和系统接口关联的模块：设计与实现
- [ ] TODO

