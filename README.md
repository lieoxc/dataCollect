## 源码说明
* 路由器测的气象站数据采集远吗
* 本源码编译完后，把二进制执行文件，作为openwrt的一个插件，添加到openwrt的固件编译中。
* 基于本项目的openwrt插件库：https://github.com/lieoxc/openwrt-package， 建议frok到自己仓库，然后clone到本地
## 源码编译（Windows环境下）

1. 下载 Golang ：https://go.dev/doc/install
2. 下载源码：
   
    ```bash
    https://github.com/lieoxc/dataCollect  // 最好frok本项目到自己仓库，然后把链接修改为自己仓库的链接
    ```
2. 进入iot-backend-router目录，并编译源码

    ```bash
    build.bat
    ```
    执行上述操作后，会在当前目录生成一个data_collect的二进制执行文件
3. 文件拷贝
  
    把iot二进制文件拷贝到：openwrt-package\data_collect\bin 目录下，替换原文件
