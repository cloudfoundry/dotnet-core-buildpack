Imports System.IO
Imports Microsoft.AspNetCore.Hosting
Imports Microsoft.AspNetCore.Builder

Namespace VBSample
  Module Program
    Sub Main(args As String())
        Dim builder as new WebHostBuilder()
        builder.UseKestrel()
        builder.UseStartup(Of Startup)
        builder.UseContentRoot(Directory.GetCurrentDirectory())
        Dim host as IWebHost
        host = builder.Build()
        host.Run()
    End Sub
  End Module

  Class Startup
    Sub Configure(app as IApplicationBuilder)
      app.UseDefaultFiles()
      app.UseStaticFiles()
    End Sub
  End Class
End Namespace