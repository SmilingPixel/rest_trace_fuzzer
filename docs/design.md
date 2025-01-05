[WIP] 技术方案 trace_rest_test



# 1. 背景

Repository: [GitHub - SmilingPixel/rest_trace_fuzzer](https://github.com/SmilingPixel/rest_trace_fuzzer)



# 2. 总体架构

![You can download the raw file, and edit using draw.io](architecture.svg)

总体分为3个模块

## 2.1. 预处理和解析


| Module      | Description                                      | Note                                                                 |
|-------------|--------------------------------------------------|----------------------------------------------------------------------|
| OpenAPI Parser | 解析 OpenAPI 文件，生成接口定义                      | 输入：一批 OpenAPI 文件，包括 1. 每个 service 的接口定义（包括 HTTP 和 gRPC） 2. 整个系统对外暴露的接口定义 输出：解析完成的接口，数据结构设计见后面章节 |
| ODG Parser  | 解析 OpenAPI，生成 ODG (Operation Dependency Graph) | 输入：解析完成的 OpenAPI 输出：ODG，包括 1. 整个系统对外暴露的接口之间的依赖关系 2. 每个 service 内部的接口之间的依赖关系（即对于图中的任何一条边，其两端的节点必须属于不同的 service） |

## 2.2. 数据持久化模块


| Module         | Description                                      | Note                                                                 |
|----------------|--------------------------------------------------|----------------------------------------------------------------------|
| Interfaces DB  | 存储解析完成的接口定义                              | 数据结构：见后面章节                                                   |
| ODG DB         | 存储 ODG                                           | 数据结构：见后面章节                                                   |
| Resource Pool  | 存储资源池，目前主要用于存储测试过程中创建的资源       | TODO: 预计参考 foREST 的设计，但这里需要为每个服务都做对应的区分         |
| Seed Queue     | 存储种子测试用例，优先级队列                         |                                                                      |


## 2.3. 测试执行模块

| Module               | Description                                      | Note                                                                 |
|----------------------|--------------------------------------------------|----------------------------------------------------------------------|
| Operation Scheduler  | 根据 ODG，调度接口的执行顺序                        | 输入：数据持久化模块中的数据 输出：本次预计执行的接口或者接口序列         |
| Operation Instantiator | 根据接口定义，实例化接口                           | 输入：接口定义，资源池数据 输出：接口请求实例                           |
| Test Driver          | 执行接口请求，记录执行结果，处理反馈，更新数据持久化模块中的数据 |                                                                      |
| Trace Analyzer       | 查询相关的 trace 数据，生成预处理反馈，引导测试        |                                                                      |




