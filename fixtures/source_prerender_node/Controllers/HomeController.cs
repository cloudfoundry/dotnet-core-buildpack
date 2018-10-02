using Microsoft.AspNetCore.Mvc;
using System.Threading.Tasks;
using Microsoft.AspNetCore.NodeServices;

namespace HelloMvc
{
    public class HomeController : Controller
    {
        [HttpGet("/")]
        public async Task<string> AddNumbers([FromServices] INodeServices nodeServices)
        {
            var result = await nodeServices.InvokeAsync<int>("./add_numbers", 1, 2);
            return "1 + 2 = " + result;
        }
    }
}
