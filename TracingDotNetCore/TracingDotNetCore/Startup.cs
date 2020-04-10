using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Datadog.Trace.OpenTracing;
using Jaeger;
using Jaeger.Reporters;
using Jaeger.Samplers;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using OpenTracing;
using OpenTracing.Util;
using TracingDotNetCore.Config;

namespace TracingDotNetCore
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        public void ConfigureServices(IServiceCollection services)
        {
            services.AddHttpClient();
            var tracingOptions = Configuration.GetSection(nameof(TracingOptions));
            var tracingSettings = tracingOptions.Get<TracingOptions>();
            if (tracingSettings.TracerTarget.Equals("Jaeger", StringComparison.OrdinalIgnoreCase))
            {
                services.AddJaegerTracer(tracingSettings, "TracingDotNetCore");
                Console.WriteLine($"JaegerHost: {tracingSettings.JaegerHost}");
                Console.WriteLine($"JaegerPort: {tracingSettings.JaegerPort}");
            }
            else if (tracingSettings.TracerTarget.Equals("DataDog",StringComparison.OrdinalIgnoreCase))
            {
                services.AddDataDogTracer();
            }
            if (tracingSettings.EnableOpenTracingAutoTracing)
            {
                services.AddOpenTracing();
                
            }
            else if (tracingSettings.EnableDataDogEnableAutoTracing)
            {
                // don't know if can be set at this point, maybe only via env variables before process start
            }

            services.AddControllers(); 
        }

        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            //if (env.IsDevelopment())
            //{
            //    app.UseDeveloperExceptionPage();
            //}

            app.UseRouting();

            app.UseAuthorization();

            app.UseEndpoints(endpoints =>
            {
                endpoints.MapControllers();
            });
        }


    }

    public static class TracerSelectionExtensions
    {
        public static IServiceCollection AddDataDogTracer(this IServiceCollection services)
        {
            ITracer tracer = OpenTracingTracerFactory.WrapTracer(Datadog.Trace.Tracer.Instance);
            GlobalTracer.Register(tracer);
            services.TryAddSingleton(tracer);
            return services;
        }

        public static IServiceCollection AddJaegerTracer(this IServiceCollection services, TracingOptions tracingOptions, string serviceName)
        {


            //string serviceName = Configuration..serServiceProvider.GetRequiredService<IWebHostEnvironment>().ApplicationName;
            var udpSender = new Jaeger.Senders.UdpSender(tracingOptions.JaegerHost, tracingOptions.JaegerPort, 0);
            var reporter = new RemoteReporter.Builder()
                .WithSender(udpSender)                  // optional, defaults to UdpSender("localhost", 6831, 0)
                .Build();
            var tracer = new Tracer.Builder(serviceName)
                .WithReporter(reporter)
                .WithSampler(new ConstSampler(true))
                .Build();



            // Allows code that can't use DI to also access the tracer.
            GlobalTracer.Register(tracer);
            // Adds the Jaeger Tracer.
            return services.AddSingleton<ITracer>(tracer);
            
        }
    }
}
