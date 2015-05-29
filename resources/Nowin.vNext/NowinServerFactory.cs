// ASP.NET 5 Buildpack
// Copyright 2014-2015 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

using System;
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;
using Microsoft.AspNet.Builder;
using Microsoft.AspNet.FeatureModel;
using Microsoft.AspNet.Hosting.Server;
using Microsoft.AspNet.Owin;
using Microsoft.Framework.ConfigurationModel;
using Nowin;
 
namespace Nowin.vNext
{
    public class NowinServerFactory : IServerFactory
    {
        private Func<IFeatureCollection, Task> _callback;

        private Task HandleRequest(IDictionary<string, object> env)
        {
            return _callback(new OwinFeatureCollection(env));
        }

        public IServerInformation Initialize(IConfiguration configuration)
        {
            var port = Environment.GetEnvironmentVariable("PORT");
            var builder = ServerBuilder.New()
                                       .SetAddress(IPAddress.Any)
                                       .SetPort(int.Parse(port))
                                       .SetOwinApp(OwinWebSocketAcceptAdapter.AdaptWebSockets(HandleRequest));

            return new NowinServerInformation(builder);
        }

        public IDisposable Start(IServerInformation serverInformation, Func<IFeatureCollection, Task> application)
        {
            var information = (NowinServerInformation)serverInformation;
            _callback = application;
            INowinServer server = information.Builder.Build();
            server.Start();
            return server;
        }

        private class NowinServerInformation : IServerInformation
        {
            public NowinServerInformation(ServerBuilder builder)
            {
                Builder = builder;
            }

            public ServerBuilder Builder { get; private set; }

            public string Name
            {
                get
                {
                    return "Nowin";
                }
            }
        }
    }
}
