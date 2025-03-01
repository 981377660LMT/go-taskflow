# GoTaskflow 框架架构与实现原理

GoTaskflow 是一个 Go 语言实现的任务流框架，灵感来源于 C++ 的 taskflow 库。它提供了一种基于有向无环图(DAG)的方式来编排和调度任务，支持并发执行、任务优先级、条件执行和嵌套子流程等特性。

## 核心组件

### 1. TaskFlow

TaskFlow 是框架的入口类，代表一组以 DAG 方式组织的任务：

```go
type TaskFlow struct {
    name  string
    graph *eGraph
}
```

- `Push` 方法将任务添加到内部图结构
- `Reset` 方法重置任务流状态

### 2. Task

Task 是基本执行单元，支持多种类型：

```go
// 创建普通任务
task := gotaskflow.NewTask("name", func() { /* 业务逻辑 */ })

// 创建条件任务
cond := gotaskflow.NewCondition("condition", func() uint { return 0 })

// 创建子流程
subflow := gotaskflow.NewSubflow("subflow", func(sf *gotaskflow.Subflow) {
    // 子流程定义...
})
```

任务依赖关系通过 `Precede` 和 `Succeed` 方法建立。

### 3. Executor

Executor 负责调度和执行任务流：

```go
type innerExecutorImpl struct {
    concurrency uint             // 最大并发数
    pool        *utils.Copool    // 协程池
    wq          *utils.Queue[*innerNode]  // 工作队列
    wg          *sync.WaitGroup  // 等待组
    profiler    *profiler        // 性能分析器
}
```

### 4. Graph 和 Node

Graph 和 Node 是内部实现，表示任务依赖关系：

```go
type eGraph struct { // 执行图
    name        string
    nodes       []*innerNode
    joinCounter *utils.RC        // 引用计数，用于跟踪未完成任务数
    entries     []*innerNode     // 入口节点(无前置依赖)
    scheCond    *sync.Cond       // 调度条件变量
    canceled    atomic.Bool      // 取消标志
}
```

## 执行流程

1. **任务创建与依赖设置**：

   - 创建任务并设置优先级
   - 通过 `Precede/Succeed` 方法建立依赖关系

2. **执行过程**：

   - 基于拓扑排序原理调度任务
   - 入度为0的节点首先执行
   - 任务完成后减少后续任务的依赖计数
   - 按优先级调度就绪任务

3. **异常处理**：
   - 捕获任务中的 panic
   - 支持子流程的异常传播

## 特色功能

### 任务优先级

任务优先级 控制就绪任务的执行顺序：

```go
task.Priority(gotaskflow.HIGH)   // 高优先级任务
```

实现上使用

slices.SortFunc

对候选任务进行优先级排序。

### 条件分支与循环

条件任务返回整数值决定执行路径，可用于实现分支和循环：

```go
cond := gotaskflow.NewCondition("cond", func() uint {
    if condition {
        return 0  // 选择第一个后继
    } else {
        return 1  // 选择第二个后继
    }
})
```

### 并发控制

基于自定义协程池

Copool

实现并发控制：

```go
type Copool struct {
    panicHandler func(*context.Context, interface{})
    cap          uint
    taskQ        *Queue[*cotask]
    corun        atomic.Int32
    coworker     atomic.Int32
    mu           *sync.Mutex
    taskObjPool  *ObjectPool[*cotask]
}
```

### 性能分析与可视化

Profiler 记录任务执行性能数据，Visualizer 提供可视化能力：

```go
// 输出性能分析数据
executor.Profile(os.Stdout)

// 生成可视化图
gotaskflow.Visualize(tf, os.Stdout)
```

## 总结

GoTaskflow 通过精心设计的组件实现了高效、灵活的任务调度框架：

1. 基于 DAG 的任务依赖管理
2. 协程池控制的高效并发执行
3. 多类型任务支持（静态、条件、子流程）
4. 完善的优先级调度机制
5. 内置的性能分析和可视化工具

这些特性使其适合于复杂依赖管理、数据处理管道和各种需要并行计算的场景。

---
