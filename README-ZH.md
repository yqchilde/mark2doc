# License部署手册

## 部署前提(必要)

1. 部署License的机器系统时钟，必须配置好时区并同步当前时间

   * 同步时区：执行`tzselect`命令，选择`4 Asia` -> `11 China` -> `1 Beijing Time` -> `1 Yes`
   * 支持ntp的使用`timedatectl set-ntp true`同步
   * 不支持ntp的使用 `timedatectl set-time "YYYY-MM-DD hh:mm:ss"` 手动设置时间

2. 对外开放端口：20088、16890 （用于给其他服务进行验证）

## 配置更新

暂无

## 1. License安装流程

<details>
<summary>点击查看License版本：</summary>


| 版本号         | 版本内容                                         | 其他依赖镜像                            |
|---------------| ----------- |----------------------------------------------|-----------------------------------|
| **V25.7.100** | - 普通用户不显示”创建人“搜索框<br/> - 修复根据”证书状态“筛选总数不对的问题 | sapphire.iam:25.6.2<br/>mongo:8.0 |
| **V25.7.101** | - 适配iam25.6.3<br/> - 配置文件增加mtls-enabled      | sapphire.iam:25.6.3<br/>mongo:8.0 |
| **v25.7.102** | - 支持数据库配置为FileDB<br/> - 修复证书列表排序问题<br> - 修复不能刷新token的问题     | sapphire.iam:25.6.4<br/>mongo:8.0（配置为FileDB时不依赖该镜像）|
| **v25.7.103** | - 修复刷新token失败的问题<br/> - 修复查看证书详情“应用限制”数不对的问题     | sapphire.iam:25.6.4<br/>mongo:8.0（配置为FileDB时不依赖该镜像）|
| **v25.7.104** | - 修复证书列表分页排序问题<br> - 修复激活时返回的“应用限制”数不对的问题     | sapphire.iam:25.6.4<br/>mongo:8.0（配置为FileDB时不依赖该镜像）|
| **v25.7.105** | - 优化激活证书过期时的返回<br>    | sapphire.iam:25.6.4<br/>mongo:8.0（配置为FileDB时不依赖该镜像）|
| **v25.7.106** | - 适配iam25.6.5<br>    | sapphire.iam:25.6.5<br/>mongo:8.0（配置为FileDB时不依赖该镜像）|
| **v25.7.107** | - 修复重启容器会重新进入引导页的问题<br> |  sapphire.iam:25.6.5<br/>mongo:8.0（配置为FileDB时不依赖该镜像）|

</details>

### 1.1 准备工作

1、请在安装License机器上提前准备好`docker`环境，并修改`/etc/docker/daemon.json`文件（如果不存在，请手动创建），增加以下内容：

```shell
{
  "insecure-registries": ["182.92.162.123:18080", "registry.amianetworks.com.cn"]
}
```

2、 使用 `systemctl restart docker` 重启docker

3、`docker`重启完成后，在`License`服务器上使用`docker login`命令登录3个私有镜像库。例如`docker login registry.amianetworks.com.cn`并输入对应的账号密码

（1）登录私有库`182.92.162.123:18080`，账号`admin`，密码`Harbor12345`

（2）登录私有库 `registry.amiasys.com` ，账号是 `amia` ，密码是`2022@Amiasys`

（3）登录私有库 `registry.amianetworks.com.cn` ，账号是 `amia` ，密码是`AmiaNetworks`

### 1.2 安装步骤

1、在服务器上新建安装目录

使用命令`vim /etc/hosts`新增一条host，内容为 `47.93.161.66 amianetworks.internal`，（如果已经添加则忽略）

2、在服务器上新建安装目录

```shell
mkdir -p asn-license
```

3、进入到安装目录，并拉取`license`最新安装配置

```shell
cd asn-license

# 配置文件支持配置mongo和FileDB两种数据库，可以自己选择配置哪种数据库，只需拉取对应的配置文件即可
# 选择一、使用mongo数据库
# 1.命令拉取config文件夹
wget -r -np -nH --cut-dirs=2 -R index.html.tmp -R index.html http://amianetworks.internal:19067/license/manager/config/

# 2.命令行拉取licensed.yml文件
wget http://amianetworks.internal:19067/license/manager/licensed.yml

# 选择二、使用FileDB数据库
# 1.命令拉取config文件夹
wget -r -np -nH --cut-dirs=2 -R index.html.tmp -R index.html http://amianetworks.internal:19067/license/filedb_manager/config/

# 2.命令行拉取licensed.yml文件
wget http://amianetworks.internal:19067/license/filedb_manager/licensed.yml
```

4、生成IAM的证书文件

```shell
# 进入asn-license目录
mkdir cert && cd cert
```

复制以下命令执行以生成证书

```shell
#=== Generate Server CA ===
openssl genrsa -out server-ca.key 4096
openssl req -x509 -new -nodes -key server-ca.key -sha256 -days 3650 -out server-ca.crt -subj "/C=US/ST=California/L=Palo Alto/O=AmiaNetworks/OU=SapphireIAM/CN=TestServer"

#=== Generate Server Cert ===
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/C=US/ST=California/L=Palo Alto/O=AmiaNetworks/OU=SapphireIAM/CN=TestServer"

cat > v3.ext <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS = localhost
EOF

openssl x509 -req -in server.csr -CA server-ca.crt -CAkey server-ca.key -CAcreateserial -out server.pem -days 825 -sha256 -extfile v3.ext

#=== Generate Client CA ===
openssl genrsa -out client-ca.key 4096
openssl req -x509 -new -nodes -key client-ca.key -sha256 -days 3650 -out client-ca.crt -subj "/C=US/ST=California/L=Palo Alto/O=Amiasys Corporation/OU=R&D/CN=iam.amiasys.com/emailAddress=contact@amiasys.com"


#=== Generate Client Cert ===
openssl genrsa -out client-license.key 2048
openssl req -new -key client-license.key -out client-license.csr -subj "/C=US/ST=California/L=Palo Alto/O=Amiasys Corporation/OU=R&D/CN=license/emailAddress=contact@amiasys.com"
openssl x509 -req -in client-license.csr -CA client-ca.crt -CAkey client-ca.key -CAcreateserial -out client-license.pem -days 825 -sha256
```

5、回到`asn-license`目录修改docker-compose配置文件

- 修改`licensed.yml`，更新`sapphire-iam`和`licensed`的镜像版本号 

<img src="./assets/license-version.png" style="width: 50%; height: auto;">

  ```yaml
  # Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.
  
  services:
    #省略其他配置...
      sapphire-iam:
      image: registry.amiasys.com/sapphire.iam:25.6.5  # 修改iam服务的版本号
      container_name: sapphire-iam
      privileged: true
      restart: always
      depends_on:
        - "asn-mdb"
      ports:
        - "17930:17930"
      volumes:
        - ./config/:/etc/sapphire/config/
        - ./cert/:/etc/sapphire/cert/
        - ./log/iam:/var/log/sapphire/
        - /etc/localtime:/etc/localtime
      networks:
        - license_network
  
    asn-licensed:
      image: registry.amianetworks.com.cn/asn-licensed:v25.7.107  # 修改license服务的版本号
      container_name: asn-licensed
      restart: always   # auto restart the container if it fails
      ports:
        - "20088:20088"
        - "16890:16890"
      depends_on:
        - "asn-mdb"
        - "sapphire-iam"
      volumes:
        - ./cert/:/etc/license/cert/
        - ./config:/etc/license/config/
        - ./log/license:/var/log/license/
        - /etc/localtime:/etc/localtime
      networks:
        - license_network
    #省略其他配置...
  
  ```

6、启动`License`容器

```shell
cd asn-license
docker compose -f licensed.yml up -d
```

7、查看License容器运行状态

```shell
docker ps -l
```

> 如果`STATUS`显示非`Up`状态，联系安装部署人员进行错误排查。

8、`License`进行更新时，先删除原来的`license`容器，然后修改容器镜像版本(**如果版本号未变，但镜像更新过，只需要`docker rm`删除本地原来的镜像即可**)，再重新启动容器

```shell
# 1.删除原来的license容器
docker compose -f licensed.yml down

# 2.修改licensed.yml文件，将asn-licensed的镜像版本修改为需要安装的版本

# 3.重新安装镜像
docker compose -f licensed.yml up -d
```

9、卸载License

(1) 仅卸载容器，保留历史数据，删除容器即可

```shell
docker compose -f licensed.yml down
```

(2) 纯净卸载，不保留历史数据

```shell
docker compose -f licensed.yml down
# 删除数据卷
docker volume rm license_mongo_data
# 删除安装目录和安装数据
rm -rf asn-license
```

### 1.3 Web管理平台

浏览器访问 `http://<ip>:16890`，按照引导程序进行配置管理员后登录

<img src="./assets/license-guide.png" style="width: 60%; height: auto;">

## 2. 开发接入License

<strong style="color: red;">步骤2 仅需开发人员阅览，测试人员请直接[跳到步骤3进行证书激活](#3-license激活验证)</strong>

### 2.1 gRPC接入

```protobuf
// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

syntax = "proto3";

option go_package = "./;license";

package license;


service LicenseManager {
  // Activate or validate a license
  rpc ActivateOrValidateLicense (License.ActivateRequest) returns (License.ActivateResponse);
}


// License related messages
message License {
  // Fixed field definitions
  message Application {
    message SwanFields {
      int32 swan_max_service_nodes = 1;   // the maximum number of service nodes
      int32 swan_max_clients = 2;         // the maximum number of clients
    }

    message ScarletteFields {
      int32 scar_max_monitoring_clients = 1; // the maximum number of monitoring clients
    }
  }

 
  message ActivateRequest {
    string license_key = 1;  // license key
    string device_id = 2;    // device id
  }

  // activate/validate license response
  message ActivateResponse {
    string license_key = 1;     // license key
    string application = 2;     // application name
    string license_type = 3;    // license type
    string create_time = 4;     // create time
    string expiration_time = 5; // expiration time
    string status = 6;          // status: active, inactive, expired
    string device_id = 7;       // device id

    // custom fields
    oneof custom_fields {
      Application.SwanFields      swan = 10; // SWAN-specific fields
      Application.ScarletteFields scarlette = 11; // Scarlette-specific fields
    }
  }
}

```

### 2.2 RESTful接入

发送请求

```shell
curl --location --request POST 'http://<ip>:16890/license/activate' \
--header 'Content-Type: application/json' \
--data-raw '{
    "licenseKey": "激活码",
    "deviceId": "设备唯一标识"
}'
```

响应返回证书信息如下：不同应用的激活返回的内容不同，目前是swan/scarlette：

```json
{
	"errCode": 0,
	"errMsg": "",
	"resp": {
		"licenseKey": "38ddc9d1780e",
		"application": "swan",
		"licenseType": "trial",
		"createTime": "2025-06-27 14:02:53",
		"expirationTime": "2025-07-04 14:02:28",
		"deviceId": "E10ADC3949BA59ABBE56E057F20F883E",
		"status": "active",
		"swan": {
			"swanMaxServiceNodes": 100,
			"swanMaxClients": 100
		}
	}
}
```

## 3. License激活验证

### 3.1 swan证书激活

step1: 通过license服务的web管理平台来创建swan的证书获得“激活码”。

<img src="./assets/license-swan-activate.png" style="width: 60%; height: auto;">

step2：然后在SWAN引导页的注册码输入框填入创建证书后获得的“激活码”来激活。

License界面变化如下：

<img src="./assets/license-activate.jpg" style="width: 60%; height: auto;">