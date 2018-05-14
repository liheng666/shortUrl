# Go语言实现一个短链接服务器

## 主要模块
![架构图](http://o99lnabej.bkt.clouddn.com/%E7%9F%AD%E9%93%BE%E6%8E%A5.PNG)
### 1. web服务器
web服务器使用Go 官方库 net/http
### 2. 短链接生成/解码 模块
主要功能是1. 输入原网址，输出短链接；2.输入短链接，输出原网址。
> 每个原网址对应一个uint64类型的ID，把ID转换成64进制的6位字符串。所以理论上能生成的短链接是2^64 (:。
>
> 考虑到大并发的情况下mysql写入会成为性能瓶颈，使用唯一ID生成器，将模块产生的数据(原链接/ID/短链接）放入消息队列，并直接向用户返回生成短链接。

### 3. 消息队列
使用通道来实现消息队列，用来存储【短链接生成模块】生成的数据

### 4. 数据存储模块

使用mysql分表存储存储 （原链接/ID/短链接）等数据


##

## 唯一ID生成器模块实现

哈哈哈，刚开始使用channel实现的方法性能并不好，直接被原子方法实现和加锁实现碾压，看来channel并不是那么的"高效"！！

```
goos: darwin
goarch: amd64
pkg: shortUrl/uuid
10000000	       195 ns/op   // channel 实现
100000000	        13.7 ns/op // 使用原子方法实现
50000000	        30.7 ns/op // 使用加锁
PASS

```

为了保证ID生成器关闭时，当前进度能正确保存，选择加锁实现方式

```
goos: darwin
goarch: amd64
pkg: shortUrl/tools
20000000	        65.1 ns/op
PASS
coverage: 19.4% of statements
```

## 短链接生成模块

### 功能
    编码： 每次请求的原链接都对应一个唯一ID（唯一ID生成器模块），将ID转码成64进制的6位字符
    解码： 将64进制的6位字符解码成唯一数字ID

### 添加单元测试（：
```
=== RUN   TestEncode
--- PASS: TestEncode (0.00s)
=== RUN   TestDecode
--- PASS: TestDecode (0.00s)
PASS
coverage: 93.1% of statements
ok      shortUrl/shortcode      0.006s
```

## 消息队列

使用缓存channel来实现消息队列FIFO



