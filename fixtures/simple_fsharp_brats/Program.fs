namespace tmp

open System
open System.IO
open Microsoft.AspNetCore.Hosting

module Program =
    let exitCode = 0

    [<EntryPoint>]
    let main args =
        let host = 
            WebHostBuilder()
                .UseContentRoot(Directory.GetCurrentDirectory())
                .UseKestrel()
                .UseIISIntegration()
                .UseStartup<Startup>()
                .Build()

        host.Run()

        exitCode