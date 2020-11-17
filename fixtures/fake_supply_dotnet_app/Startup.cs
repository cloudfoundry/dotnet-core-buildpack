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
                var bosh2 = new Process();
                bosh2.StartInfo.FileName = "bosh2";
                bosh2.StartInfo.Arguments = "--version";
                bosh2.StartInfo.RedirectStandardOutput = true;
                bosh2.Start();

                var output = bosh2.StandardOutput.ReadToEnd();
                bosh2.WaitForExit();
                return context.Response.WriteAsync("bosh2: " + output);
            });
        }
    }
}
