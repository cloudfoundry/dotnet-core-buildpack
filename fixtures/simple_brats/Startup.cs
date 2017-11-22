using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;

namespace HelloWeb
{
    public class Startup
    {
        public void Configure(IApplicationBuilder app)
        {
            app.Run(context =>
            {
                if (context.Request.Path == "/") {
                    return context.Response.WriteAsync("Hello World! ");
                } else {
                    context.Response.StatusCode = 404;
                    return context.Response.WriteAsync("404 Not Found");
                }
            });
        }
    }
}
