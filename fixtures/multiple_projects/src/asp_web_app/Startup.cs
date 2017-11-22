using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using console_app;

namespace HelloWeb
{
    public class Startup
    {
        public void Configure(IApplicationBuilder app)
        {
            app.Run(context =>
            {
                return context.Response.WriteAsync(UseProjectTwo());
            });
        }
        
        public string UseProjectTwo()
        {
            return ProjectTwo.GetAString();
        }
    }
}
