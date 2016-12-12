open System
open Microsoft.AspNetCore.Builder
open Microsoft.AspNetCore.Hosting
open Microsoft.Extensions.Configuration

type Startup() =
    member this.Configure(app: IApplicationBuilder) = 
      app.UseDefaultFiles() |>ignore
      app.UseStaticFiles() |>ignore


[<EntryPoint>]
let main argv = 
    let config = ConfigurationBuilder().AddCommandLine(argv).Build()

    let host = WebHostBuilder().UseKestrel().UseConfiguration(config).UseStartup<Startup>().Build()
    host.Run()
    0
