# libbuildpack-sealights
Cloud Foundry Buildpack integrations with Sealights

## Bind your application to your Sealights service

1. First step is to create and configure user provided service

    For Linux:
    ```cf cups sealights -p '{"token":"ey…"}'```
    For Windows:
    ```cf cups sealights -p "{\"token\":\"ey…\"}"```

    Note: you can change prameters later with command `cf uups sealights -p ...`

    Complete list of the prameters currently supported by the buildpack service is:
    ```
    {
        "version": "string"                 //sealights version. default value - latest
        "customAgentUrl": "string"          //sealights agent will be downloaded from this url if provided
        "mode": "string"                    //execution stage. values: [config, scan, startExecution, testListener, endExecution]. default value - testListener
        "token": "string"                   //sealights token
        "tokenFile": "string"               //sealights token filename
        "bsId": "string"                    //sealights session id
        "bsIdFile": "string"                //sealights session id filename
        "target": "string"                  //[testListener] target. default value is calculated based on app start command 
        "targetArgs": "string"              //[testListener] target args. default value is calculated based on app start command 
        "workingDir": "string"              //[testListener] target working directory. default value is calculated by cf
        "profilerLogDir": "string"          //[testListener] profiler will write logs to this directory
        "profilerLogLevel": "string"        //[testListener] profiler log level
        "labId": "string"                   //lab id
        "proxy": "string"
        "proxyUsername": "string"
        "proxyPassword": "string"
        "ignoreCertificateErrors": "string" 
        "tools": "string"
        "tags": "string"
        "notCli": "string"
        "appName": "string"                 //[config] application name
        "branchName": "string"              //[config] branch name
        "buildName": "string"               //[config] build number
        "includeNamespace": "string"        //[config] namespaces allow list
        "workspacePath": "string"           //[scan] folder to scan. default value is app directory
        "ignoreGeneratedCode": "string"     //[scan] ignore generated code
        "testStage": "string"               //[startExecution] test stage name
    }
    ```

2. Bind your application to your Sealights service

    cf bind-service [app name] sealights

