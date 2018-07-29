首发于：https://studygolang.com/articles/13515

# Golang 下的微服务 - 第 9 部分 - 使用 CircleCI 部署

欢迎大家在 [Patreon](https://www.patreon.com/ewanvalentine) 上向我提供更多诸如此类的素材。

在本系列的这一章节，我们将简要介绍使用 [CircleCI](http://circleci.com/) 与我们的其中一项服务建立持续集成。

[CircleCI](http://circleci.com/) 是一款不可思议的工具，它有一个非常实用的免费平台。这个平台就是 SaaS, 因此与 Jenkins 不同的是，它是被完全管理的。同时它的配置和建立非常直截了当。此外，[CircleCI](http://circleci.com/) 也使用 Docker 镜像（images)，所以你可以在如何管理你的构建上有很多创意。

确保你已经注册并且创造了账户。首先让我们在 [CircleCI](http://circleci.com/) 中新建一个工程。在左侧菜单中，点击 “add project”。如果你已经将你的 github 账户连接到你的 [CircleCI](http://circleci.com/) 账户，你应该可以看到你的微服务 git 仓库出现在列表中。点击 “follow project”。你将看到一个请求页面，你可以选择你乐于使用的操作系统和语言。确保 Linux 和 Go 被选中。然后点击开始构建。

这将创造一些默认的配置，但是我们需要在构建能开始正常工作之前，增加我们自己的配置到此代码仓库中。

所以在我们的服务中（我将为此使用我们的委托服务），在项目根目录新建一个文件夹 `mkdir .circleci`，然后在此文件夹下新建一个文件 `touch .circleci/config.yml`。现在让我们开始增加我们的构建配置。

```
version: 2
jobs:
  build:
    working_directory: /app
    docker:
      # Here we select the Docker images we wish to use
      # in order to build our service.
      # We're using a container I made, which configurs the Google Cloud SDK
      # Kubernetes, and a few other utils. This is open-source, and you can find
      # the repo here https://github.com/EwanValentine/gcloud-docker-kubernetes
      #
      # Then we're using the docker image itself, so that we can build docker containers.
      - image: ewanvalentine/gcdeploy:latest
        environment:
          GCLOUD_PROJECT_NAME: shippy-freight
          GCLOUD_CLUSTER_NAME: shippy-freight-cluster
          CLOUDSDK_COMPUTE_ZONE: europe-west2-a

          # This is a google service key, which allows us to authenticate
          # our build process with our cluster.
          # You need to generate a service key, such as the one we generated
          # in part 7. You can copy the contents of this and encode it using base64.
          # Then add the base64 string into your environment variables, in the settings
          # of this build project. To find this, click on the spanner icon in your build.
          # Then click on environment variables, click add variable, with the name GCLOUD_SERVICE_KEY
          # then paste the base64 string of your service key into the value and save that.
          GOOGLE_APPLICATION_CREDENTIALS: ${HOME}/gcloud-service-key.json

      - image: docker:17.05.0-ce-git
        environment:
          DOCKER_TAG_PREFIX: "eu.gcr.io/$GCLOUD_PROJECT_NAME/shippy-consignment-service"
          DOCKER_TAG: "$DOCKER_TAG_PREFIX:$CIRCLE_SHA1"
    steps:
      - checkout
      - setup_remote_docker
      - run:
          name: Install dependencies

          # Fetches the base64 encoded service key content, decodes it into a file again.
          # Then sets the gcloud project name from the environment variables we set above.
          # Then we set the cluster name, the compute region/zone, then fetch the credentials.
          command: |
            echo $GCLOUD_SERVICE_KEY | base64 --decode -i > ${HOME}/gcloud-service-key.json && \
              gcloud auth activate-service-account --key-file ${HOME}/gcloud-service-key.json && \
              gcloud config set project $GCLOUD_PROJECT_NAME && \
              gcloud --quiet config set container/cluster $GCLOUD_CLUSTER_NAME && \
              gcloud config set compute/zone ${CLOUDSDK_COMPUTE_ZONE} && \
              gcloud --quiet container clusters get-credentials $GCLOUD_CLUSTER_NAME
      - deploy:
          name: Push application Docker image
          command: |
            make deploy
```

为了使之生效我们需要做一些事情，我在评论中已经谈到了这一点，但它是一个重要的步骤，所以我想重申这一部分。

我们需要谷歌云服务钥匙，正如我们在[第 7 章](https://studygolang.com/articles/12799)创建的那个，然后我们需要将此钥匙加密成 base64 并作为我们构建工程设置中的一个环境变量来存储。

因此找到你的谷歌云服务钥匙，然后运行 `$ cat <keyname>.json | base64`，并复制得到的字符串。回到 [CircleCI](http://circleci.com/) 你的项目来，点击右上方的齿轮，然后选择左边栏目中的环境变量。新建一个环境变量，命名为`GCLOUD_SERVICE_KEY`，然后粘贴前面得到的 base64 字符串作为其值，并保存。

上述操作可以在 circleci 内保存任何安全信息，且使代码仓库不熟任何敏感数据影响。它将这些访问密钥保存在操作团队的控制之下，而不仅限于任何可以访问代码仓库的人员。

现在我们的构建配置中，用来对我们组进行身份验证的变量，其内容被解码为一个文件。

大功告成，相当简单。我们拥有 CI 作为我们的一个服务。作为一个产品服务，在你执行部署步骤之前，你可能会首先运行你的测试用例。查看[这篇文档](https://circleci.com/docs/2.0/)然后看看你能用 circle 做哪些有趣的东西。由于 circle 使用Docker 容器，你甚至可以加一个数据库容器，以便于你运行集成测试。发挥你的创造力，最大限度使用这些特性。

如果你觉得这一系列的文章对你有用，且你装了广告拦截器(没有人会责怪你)，考虑打赏一下我撰写这些文章所付出的时间和努力吧！谢谢！
https://monzo.me/ewanvalentine

或者，在[Patreon](https://www.patreon.com/ewanvalentine)上向我提供更多诸如此类的素材。

---

via: https://ewanvalentine.io/microservices-in-golang-part-9/

作者：[Ewan Valentine](https://ewanvalentine.io/author/ewan/)
译者：[MAXAmbitious](https://github.com/MAXAmbitious)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
