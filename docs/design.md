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
| OpenAPI Parser | 解析 OpenAPI 文件，生成接口定义                      | 输入：一批 OpenAPI 文件，包括 1. 每个 service 的接口定义（包括 HTTP 和 gRPC） 2. 整个系统对外暴露的接口定义 输出：解析完成的接口，数据结构设计见后面章节 |
| ODG Parser  | 解析系统外部暴露 OpenAPI，生成 ODG (Operation Dependency Graph) | 输入：解析完成的 OpenAPI 输出：ODG，包括 1. 整个系统对外暴露的接口之间的依赖关系 2. 每个 service 内部的接口之间的依赖关系（即对于图中的任何一条边，其两端的节点必须属于不同的 service） |
| DFG Parser  | 内部服务 API，生成 DFG (Data Flow Graph)            | 输入：解析完成的 OpenAPI 输出：DFG，包括 1. 每个 service 内部的接口之间的数据流传递关系（即对于图中的任何一条边，其两端的节点必须属于不同的 service） |

## 2.2. 数据持久化模块


| Module         | Description                                      | Note                                                                 |
|----------------|--------------------------------------------------|----------------------------------------------------------------------|
| Interfaces DB  | 存储解析完成的接口定义                              | 数据结构：见后面章节                                                   |
| ODG DB         | 存储 ODG                                           | 数据结构：见后面章节                                                   |
| DFG DB         | 存储 DFG                                           | 数据结构：见后面章节                                                   |
| Runtime Info Graph         | 存储程序运行时的 DFG 调用和覆盖状态                                          | 数据结构：见后面章节                                                   |
| Resource Pool  | 存储资源池，目前主要用于存储测试过程中创建的资源       | TODO: 预计参考 foREST 的设计         |
| Testcase Queue     | 存储测试用例，优先级队列，依据覆盖率提升等指标计算优先级                         |                                                                      |


## 2.3. 测试执行模块

| Module               | Description                                      | Note                                                                 |
|----------------------|--------------------------------------------------|----------------------------------------------------------------------|
| Testcase Scheduler  | 根据 ODG, DFG, Runtime Info Graph 等，调度接口的执行顺序                        | 输入：数据持久化模块中的数据 输出：本次预计执行的接口序列         |
| Operation Population | 根据接口定义，填充参数，实例化接口                           | 输入：接口定义，资源池数据 输出：接口请求实例                           |
| Test Driver          | 执行接口请求，记录执行结果，处理反馈，更新数据持久化模块中的数据 |                                                                      |
| Response Checker       | 收集系统响应，存储测试结果        |                                                                      |
| Trace Analyser       | 收集 Trace，同步更新 RuntimeInfo Graph 的状态        |                                                                      |


# 3. 具体实现方案

## 3.1. OpenAPI 解析

框架选取 [getkin/kin-openapi](https://github.com/getkin/kin-openapi/)

解析时，对于输入的 OpenAPI Spec，构建两套数据：
1. 整个系统的接口定义
2. 每个 service 的接口定义，存储为一个 map，key 为 service name [TODO: 这里需要确认一下]


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
| source_resource  | [TODO: 待定] | 表示存在流动的资源（例如某个参数） |

Graph，包含一组 Edge，表示整个系统内部API之间的数据流传递关系。


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

### 3.2.3. DFG 解析

- 类似于 ODG 解析，目前设计主要依赖于资源名称的解析
- 更加侧重于“**数据流向**”的依赖关系，因此解析的资源对，要么同时来自于两个 Operation 的输入参数，要么同时来自于两个 Operation 的输出参数


#### 3.2.4. 实现差异

对于DFG，由于是内部 API 调用的依赖，因此更加侧重于“**资源流向**”的依赖关系，例如：
- CheckoutService 的 placeOrder 调用了 PaymentService 的 charge
- placeOrder 输入参数和 charge 输出参数中，都有 currency 字段
- 我们认为，这一信息，表示该数据从 CheckoutService 流向 PaymentService
- 因此，我们在 ODG 中，应该有一条边，表示 placeOrder 依赖于 charge

主要依赖于两条规则：
1. 如果两个参数的 normalized name 相同，那么认为这两个参数有依赖关系
2. 两个操作之间存在“**资源流向**”的依赖关系

## 3.3. Resource Pool


[TODO: 待定]


## 3.4. Testcase Queue

总体思路：
- ~~从 ODG 中找到入度为 0 的节点，作为种子节点~~
- 从种子节点开始，根据 ODG 的依赖关系，构建测试序列
- 将测试序列中的节点加入 Testcase Queue （类似于 DFS 的思路）
- 优先级:
  - 提高内部 ODG 覆盖率的测试用例具有更高的优先级
  - 提高系统外部接口覆盖率（Path, Status Code）的测试用例具有更高的优先级
  - ...

[TODO: 具体细节待定]


## 3.5. Test Driver

测试执行，发送请求和处理响应，更新数据持久化模块中的数据


## 3.6. Trace Analyzer

根据 trace 数据，生成预处理反馈，引导测试

### 3.6.1. Trace 搜集

- OpenTelemetry Demo 中，使用 Jaeger 提供的 API 进行 trace 搜集
- 参考这个 [issue](https://github.com/orgs/jaegertracing/discussions/2876)


### 3.6.2. Trace 分析

1. 记录覆盖的调用边和调用次数，更新 ODG 实例
2. 统计覆盖率更新，上报给 Runtime Info Graph
3. 更新 ODG，补充缺失的依赖关系，修改错误的依赖关系 [TODO: 具体实现待定]


### 3.7. Runtime Info Graph

- 记录程序运行时的 DFG 调用和覆盖状态
- 基于 DFG，目前用于记录 DFG 上的调用次数，用于后续的覆盖率统计和测试用例优先级计算

Edge， a -> b 表示存在数据流从 a 到 b，同时携带了调用次数信息
| Field            | Type       | Description                         |
|------------------|------------|-------------------------------------|
| source           | Operation  | 表示源 Operation               |
| target           | Operation  | 表示目标 Operation             |
| source_resource  | [TODO: 待定] | 表示存在流动的资源（例如某个参数） |
| hit_count       | int        | 表示调用次数                       |


# 4. 子任务排期

- [x] 初步的技术方案设计，包括整体架构和模块设计
- [x] OpenAPI 解析模块的实现
- [ ] RestTestGen 的 ODG 解析模块的移植
- [ ] TODO

