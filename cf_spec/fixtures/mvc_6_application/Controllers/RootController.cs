using Microsoft.AspNet.Mvc;
using System;
using System.Collections;

namespace Nora5.Controllers
{
    [Route("/")]
    public class RootController : Controller
    {
        [HttpGet]
        public string Get()
        {
            return "Hi, I'm Nora!\n";
        }

        [HttpGet("env")]
        public IDictionary Env()
        {
            return Environment.GetEnvironmentVariables();
        }

    }
}
