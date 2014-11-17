# NBE-Levi

## 这是干嘛的？

Levi 是 NBE 核心里面一个代理。它主要负责将 Docker 反向暴露连接 Dot 主节点，并接受 Dot 主节点的控制。每一个接入的 Docker Host 均会有一个对应的 Levi 跟它跑在一起。Levi 只能跑在 Root 权限下，同时 Levi 会实时的更新节点容器状态，汇总写入到 Etcd 中。

## 怎么运行他呢？

Levi 有一个配置文件，`levi.yaml`，运行的时候只需要使用文件路径从命令行传入即可。

    levi -c levi.yaml [-DEBUG]

`-DEBUG` 指定是否为 debug 模式, debug 模式下会输出一些 debug 信息, 如任务的格式, app.yaml 和 config.yaml 的数据内容等。

## Levi 的运行步骤

Levi 在接受任务之后会执行以下几步：

1. 压缩任务，多个任务会合并在一个时间点运行。
2. 判断任务类型，现在支持 build/test/run 3类任务。
3. 如果有 App 需要更新 Nginx，那么 Levi 会在等待这一波任务全部完成的情况下调用 ngx_dyups 动态接口更新 nginx upstream。
4. 最后 Levi 会实时的返回任务执行结果，并且使用一样的协议。

## Levi 的任务类型

1. Build

Build 类型的任务将会把代码从 Git 仓库中拉取下来，Checkout 成对应版本，并按照一定规则打包成 Docker Image 并 Push 到远端的 Docker-Registry 之上。所有的 App(Service) 必须经过这一步之后才能执行后面的步骤。项目二进制编译也是在这一步，Levi 将会执行 App.yaml 中指定的 Build 命令。

2. Test

当一个镜像打包完毕之后，Levi 可以接受这个镜像的测试任务。类似于 Drone 的行为，用户可以通过在 App.yaml 中指定的 Test 命令让 Levi 运行这个镜像。如果返回值不为0，Levi 会判断为测试失败，并返回结果。

3. Add/Remove/Update

当接收到此类任务类型的时候，Levi 会判断任务类型，并对 Docker 进行相应的操作。最终当任务执行完毕，如果有必要 Levi 将会更新 Nginx 的 upstream。Levi 支持不暴露端口的 Daemon 型任务。
