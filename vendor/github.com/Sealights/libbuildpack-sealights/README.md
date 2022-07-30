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
        "version"               // sealights version. default value - latest
        "verb"                  // execution stage. values: [config, scan, startExecution, testListener, endExecution]
                                // in case if stage is not provided sealights service will not be called on container start
        "customAgentUrl"        // sealights agent will be downloaded from this url if provided
        "customCommand"         // allow to replace application start command
        "labId"                 // will be downloaded agent version of the specified lab
        "proxy"                 // proxy for the agent download client
        "proxyUsername"         // proxy user
        "proxyPassword"         // proxy password
        "enableProfilerLogs"    // allow to enable logs in the profiler when listener is started in the background mode

        + rest of the arguments that required for sealights service
    }
    ```

2. Bind your application to your Sealights service

    cf bind-service [app name] sealights

3. Restage an application to apply the changes

    cf restage [app name]

