1. 项目分为四个目录
 - server 聊天服务器代码
   - 使用 session 管理每一个连接会话
   - session 使用状态机管理当前 session 所处的状态，共三个state：threshold、lobby、channel
     - threshold 状态： 接受客户端登陆请求
     - lobby 状态： 接受客户端进入指定房间请求
     - channel 状态：接受客户端聊天与退出房间请求
   - UserManager 用户信息管理
   - ChannelManager 聊天房间（频道）管理
      - 历史聊天记录使用循环数组，去除内存搬移操作
   - 脏字过滤功能（未实现）
   - GM指令与用户在线时长统计（未实现）
   - 单元测试（未使用过 golang 单元测试）
 - client 客户端代码
 - utils 辅助库
 - common 服务器与客户端共用代码，放置协议文件等
 - tools 工具
    - protoc protobuf 代码生成器
    - build.sh 编译脚本
    
2. 使用说明
 - 使用 tools/build.sh 编译工程，二进制文件生成在 bin/ 目录
 - 执行 ./bin/server 启动服务器
 - 执行 ./bin/client 启动客户端
    - 进入 Threshold 状态时，输入指令登陆：login <用户名>
    - 进入 Lobby 状态时，输入指令进入指定房间：enter <房间名>
    - 进入 Channel 状态时
      - 输入指令聊天：say <聊天内容>
      - 输入指令退出房间（进入Lobby状态）：leave
    
3. 性能指标未测试
   
4. 如何扩展
- 用户鉴权与数据落地
- 同时进入多个房间
   
5. 使用第三方库
   - protobuf - 在服务器与客户端进行通讯
   
