# reproduce-pika-parallel-stateful-command

复现 [#935](https://github.com/OpenAtomFoundation/pika/issues/935)

根本原因是 `redis` 是单线程的，至于为啥要这样，网上有很多文章（如[这篇](https://medium.com/@jychen7/sharing-redis-single-thread-vs-multi-threads-5870bd44d153)）。

但是 `pika` 的命令执行是多线程的。

最初 issue 的关键词是 `pipeline`，而这个问题，其实跟 pipeline 无关。

跟**并发**有关。比如同时读 `key` 这个 list 下标 0~9，redis 的话，不会读到重复数据（pipeline 的第 2 个命令就给 trim 掉了），而 pika 会读到重复数据，因为同一时刻、相同下标确实内容是相同的。
