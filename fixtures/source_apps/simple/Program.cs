var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddControllersWithViews();

var app = builder.Build();

// Configure the HTTP request pipeline.
if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Home/Error");
    // The default HSTS value is 30 days. You may want to change this for production scenarios, see https://aka.ms/aspnetcore-hsts.
    app.UseHsts();
}

app.UseHttpsRedirection();
app.UseStaticFiles();

app.UseRouting();

app.UseAuthorization();

app.MapControllerRoute(
    name: "default",
    pattern: "{controller=Home}/{action=Index}/{id?}");

// Reports the GC's actual resolved GCHeapHardLimit configuration (via GC.GetConfigurationVariables(),
// available since .NET 7) rather than the raw DOTNET_GCHeapHardLimit environment variable, proving
// the CLR itself parsed and applied the value rather than merely being present in the shell environment.
// The "GCHeapHardLimit" key is always present in the dictionary (0 when unconfigured), so no fallback is needed.
app.MapGet("/env/dotnet-gc-heap-hard-limit", () => $"0x{(long)GC.GetConfigurationVariables()["GCHeapHardLimit"]:x}");

app.Run();
