using Microsoft.AspNet.Builder;
using Microsoft.AspNet.Hosting;

public class Startup
{
	public void Configure(IApplicationBuilder app)
	{
		app.UseDefaultFiles();
		app.UseStaticFiles();
    }

    public static void Main(string[] args) => WebApplication.Run(args);
}
