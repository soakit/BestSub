# 手把手教你全自动永久免费获取高质量可解锁流媒体，解锁手机 APP 版 ChatGPT 的订阅节点

仅推荐自己本地使用（部署到 NAS，软路由等设备上，后面会做一个 luci 包，现已有 docker 部署方式）  
海外机器检测出来的节点，自己本地不一定可以使用  
没时间折腾的朋友直接跳转到文末，有我维护的订阅链接

1.  首先根据自己系统下载对应的文件  
    [Windows](https://github.com/bestruirui/BestSub/releases/latest/download/BestSub_Windows_x86_64.zip) | [Linux](https://github.com/bestruirui/BestSub/releases/latest/download/BestSub_Linux_x86_64.tar.gz) | [Mac](https://github.com/bestruirui/BestSub/releases/latest/download/BestSub_Darwin_aarch64.tar.gz)
    
2.  下载配置文件  
    [config.example.yaml](https://raw.githubusercontent.com/bestruirui/BestSub/refs/heads/master/doc/config.example.zh.yaml)  
    [rename.yaml](https://raw.githubusercontent.com/bestruirui/BestSub/refs/heads/master/doc/rename.yaml)
    
3.  修改配置文件，这里给出一个修改后的 (兄弟们注意多看看 Github，更新的有点频繁，一直在完善功能，这里的更新可能不及时)
    

替换 sub-urls 后即可使用

`# 是否打印进度  print-progress: false  log-level: info  # 重命名方法 api 或 regex 或 mix rename:   method: regex   flag: true  check:   # 并发   concurrent: 100   # 检查间隔,单位分钟   interval: 30   # 超时时间 单位毫秒   timeout: 2000   # 最低测速 单位KB/s   min-speed: 2048   # 下载测试超时时间(s)   download-timeout: 10   # 测速地址   speed-test-url: https://github.com/AaronFeng753/Waifu2x-Extension-GUI/releases/download/v3.121.12-beta/Update-W2xEX-v3.121.12-beta-FROM-v3.121.01.7z   # 跳过测速的名称   speed-skip-name: 倍率|x\d+   # 测速并发   speed-check-concurrent: 10   # 检查项目   items:     # - openai     # - youtube     # - netflix     # - disney     # - speed  save:   # 保存方法 webdav 或 http 或 gist 或 r2   method:      - http   # 保存端口   port: 18989  # mihomo api mihomo-api-url: "" # mihomo api secret mihomo-api-secret: "" # 重试次数 sub-urls-retry: 3 # 代理设置 支持 http 和 socks 代理 proxy:   type: "" # 可选值: http, socks   address: "http://192.168.31.11:7890" # 代理地址 # 订阅链接 sub-urls:   - https://your-sub-url/sub1   - https://your-sub-url/sub2`
sub-urls 获取方法

1.  直接在 github 搜索即可，支持 clash，base64，v2ray 格式的  
    例如 `https://raw.githubusercontent.com/peasoft/NoMoreWalls/refs/heads/master/list.yml`  
    无法直连的自行在搜索引擎搜索 `Github加速`
2.  使用 fofa 等工具搜索 (自行在论坛搜索)  
    [https://linux.do/t/topic/201723](https://linux.do/t/topic/201723)
3.  这个帖子有分享  
    [https://linux.do/t/topic/190390](https://linux.do/t/topic/190390)

`https://raw.githubusercontent.com/snakem982/proxypool/main/source/clash-meta.yaml https://raw.githubusercontent.com/ripaojiedian/freenode/main/clash https://raw.githubusercontent.com/vxiaov/free_proxies/main/clash/clash.provider.yaml https://raw.githubusercontent.com/chengaopan/AutoMergePublicNodes/master/list.yml https://fs.v2rayse.com/share/20240725/mbm257hpyj.yaml https://raw.githubusercontent.com/Ruk1ng001/freeSub/main/clash_top30.yaml`
(可选) 检测结束后自动更新对应客户端的订阅

[![{B6EB274E-0990-430F-B92B-A605C85C7E29}](https://cdn3.ldstatic.com/optimized/4X/e/5/6/e562515445a5946033df45e150b43480ad19a609_2_690x475.png)](https://cdn3.ldstatic.com/original/4X/e/5/6/e562515445a5946033df45e150b43480ad19a609.png "{B6EB274E-0990-430F-B92B-A605C85C7E29}")

  
将这里的外部控制地址填入到配置文件中

`# mihomo api mihomo-api-url: "http://127.0.0.1:9090" # mihomo api secret mihomo-api-secret: ""`

5.  重命名 `confg.example.yaml` 为 `confg.yaml`
6.  检查目录结构

`bestsub/ │ ├── config/ │   ├── config.yaml │   └── rename.yaml │ ├── BestSub.exe`

[![{7FC9327D-A66A-42AF-A75E-9E9752D9B9F3}](https://cdn3.ldstatic.com/optimized/4X/3/9/5/3955fb9729d9a00e920cb4cae3573753c9ae7b53_2_562x500.jpeg)](https://cdn3.ldstatic.com/original/4X/3/9/5/3955fb9729d9a00e920cb4cae3573753c9ae7b53.jpeg "{7FC9327D-A66A-42AF-A75E-9E9752D9B9F3}")

7.  在终端运行`.\BestSub.exe`，如果使用我上面给出的配置文件的话，现在就可以打开  
    `http://localhost:18989` 耐心等待结果了，一般来说，等待时间最长超不过十分钟  
    
    [![{2D6841D6-4F20-48A4-8B1A-26B7C57E7108}](https://cdn3.ldstatic.com/original/4X/b/7/1/b711372cf35c1711cc7e3785466ac715a85d518f.png)](https://cdn3.ldstatic.com/original/4X/b/7/1/b711372cf35c1711cc7e3785466ac715a85d518f.png "{2D6841D6-4F20-48A4-8B1A-26B7C57E7108}")
    
      
    
    [![{D5CF40F3-26C0-46C8-964D-4362AF63FCCF}](https://cdn3.ldstatic.com/optimized/4X/f/6/2/f62598b69c76e647ac9711f2f30e1e2565e78038_2_690x462.png)](https://cdn3.ldstatic.com/original/4X/f/6/2/f62598b69c76e647ac9711f2f30e1e2565e78038.png "{D5CF40F3-26C0-46C8-964D-4362AF63FCCF}")
    
      
    这里可以看出：近 3000 个节点，耗时 3 分钟  
    
    [![image](https://cdn3.ldstatic.com/optimized/4X/d/9/e/d9eec3497b85ec70bd2c385c58d208a7fa5094d2_2_613x500.png)](https://cdn3.ldstatic.com/original/4X/d/9/e/d9eec3497b85ec70bd2c385c58d208a7fa5094d2.png "image")
    
8.  因为我上面给出的配置文件中 `items` 项都注释掉了，所以这里仅有一个 all.yaml  
    
    [![{B590441C-13B5-4F64-85EE-46742447BCF4}](https://cdn3.ldstatic.com/optimized/4X/f/0/7/f0781df6817c0bde2df9c59663d0af81db1da986_2_690x480.png)](https://cdn3.ldstatic.com/original/4X/f/0/7/f0781df6817c0bde2df9c59663d0af81db1da986.png "{B590441C-13B5-4F64-85EE-46742447BCF4}")
    
9.  上面的各个链接都是纯节点信息，没有规则，直接导入对应的客户端无法使用，需要自行配置规则，这里给出一个我配置好的 [base.yaml](https://raw.githubusercontent.com/bestruirui/BestSub/refs/heads/master/doc/base.yaml)  
    修改对应的链接
    

`proxy-providers:   ProviderALL:     url: http://localhost:18989/all.yaml     type: http     interval: 600     proxy: DIRECT     path: ./proxy_provider/ALL.yaml   ProviderOpenai:     url: http://localhost:18989/openai.yaml     type: http     interval: 600     proxy: DIRECT     path: ./proxy_provider/Openai.yaml   ProviderYoutube:     url: https://localhost:18989/youtube.yaml     type: http     interval: 600     proxy: DIRECT     path: ./proxy_provider/Youtube.yaml   ProviderNetflix:     url: https://localhost:18989/netfilx.yaml     type: http     proxy: DIRECT     path: ./proxy_provider/Netflix.yaml   ProviderDisney:     url: https://localhost:18989/disney.yaml     type: http     interval: 600     proxy: DIRECT     path: ./proxy_provider/Disney.yaml`

10.  现在可以把 `base.yaml` 导入到对应的客户端使用了  
    
    [![{36A6E548-5668-41AE-B307-BBB17BD1D7B5}](https://cdn3.ldstatic.com/optimized/4X/9/5/0/9506fb24fedd4a3801b305daf723a74b46837291_2_681x500.png)](https://cdn3.ldstatic.com/original/4X/9/5/0/9506fb24fedd4a3801b305daf723a74b46837291.png "{36A6E548-5668-41AE-B307-BBB17BD1D7B5}")
    
      
    
    [![image](https://cdn3.ldstatic.com/optimized/4X/5/4/3/543cf9c5c662416b20a82a39313361ae0961bed0_2_690x499.png)](https://cdn3.ldstatic.com/original/4X/5/4/3/543cf9c5c662416b20a82a39313361ae0961bed0.png "image")
    

第一次写教程，看不懂的评论就好，我会一一回复的

另外 bug 反馈，功能请求请移至 github issue

[github.com](https://github.com/bestruirui/BestSub)

![](https://cdn3.ldstatic.com/optimized/4X/c/7/0/c7058f198b42b0f3c9067578803e4ecb8148a3d1_2_690x344.png)

### [GitHub - bestruirui/BestSub: Best Sub, Best for Your Net](https://github.com/bestruirui/BestSub)

Best Sub, Best for Your Net

我想要![:star:](https://linux.do/images/emoji/twemoji/star.png?v=14 ":star:")![:star:](https://linux.do/images/emoji/twemoji/star.png?v=14 ":star:")马上破千了兄弟们！  
最后放一个给没时间动手的朋友一个我自己维护的永久订阅链接，不定期更新  
下面这条订阅是带有完整分流规则的 clash/mihomo 订阅，实测 mihomo 裸核，mihomo party ，flclash 直接导入即可使用

`https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/bestsub.yaml`

下面这些订阅是 mihomo 格式的纯节点订阅链接，需要搭配对应的订阅转换使用

`# 全部节点 https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/all.yaml # 速度快的节点 https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/speed.yaml # Youtube https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/youtube.yaml # Disney https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/disney.yaml # Netflix https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/netflix.yaml # ChatGPT https://bestsub.bestrui.ggff.net/share/bestsub/cdcefaa4-1f0d-462e-ba76-627b344989f2/openai.yaml`