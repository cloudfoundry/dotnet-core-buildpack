using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using System.Diagnostics;

namespace HelloWeb
{
    public class Startup
    {
        public void Configure(IApplicationBuilder app)
        {
            app.Run(context =>
            {
                var go = new Process();
                go.StartInfo.FileName = "go";
                go.StartInfo.Arguments = "version";
                go.StartInfo.RedirectStandardOutput = true;
                go.Start();

                var output = go.StandardOutput.ReadToEnd();
                go.WaitForExit();
                return context.Response.WriteAsync("go: " + output);
            });
        }
    }
}
