using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;
using OpenTracing;

namespace TracingDotNetCore.Controllers
{
    [ApiController]
    [Route("[controller]")]
    public class WeatherForecastController : ControllerBase
    {
        private static readonly string[] Summaries = new[]
        {
            "Freezing", "Bracing", "Chilly", "Cool", "Mild", "Warm", "Balmy", "Hot", "Sweltering", "Scorching"
        };

        private readonly ILogger<WeatherForecastController> _logger;
        private readonly ITracer _tracer;
        private readonly IHttpClientFactory _httpClientFactory;

        public WeatherForecastController(ILogger<WeatherForecastController> logger, ITracer tracer, IHttpClientFactory httpClientFactory)
        {
            _logger = logger;
            _tracer = tracer;
            _httpClientFactory = httpClientFactory;
        }

        [HttpGet("logs")]
        public string GetLogs()
        {
            var sb = new StringBuilder();
            Directory.GetFiles("/var/log/datadog/dotnet/").ToList().ForEach(f =>
            {
                sb.Append(System.IO.File.ReadAllText(f));
                sb.Append(Environment.NewLine);
            });
            return sb.ToString();
        }
        [HttpGet("envs")]
        public string GetEnv()
        {
            var sb = new StringBuilder();
            foreach(string key in Environment.GetEnvironmentVariables().Keys) { 
                sb.AppendLine(key + "=" + Environment.GetEnvironmentVariable(key));
            };
            return sb.ToString();
        }

        [HttpGet]
        public async Task<IEnumerable<WeatherForecast>> Get()
        {
            var builder = _tracer.BuildSpan("customSpan");

            using (var scope = builder.StartActive(true))
            {
                scope.Span.Log("entro");
                var rng = new Random();
                var ret = Enumerable.Range(1, 5).Select(index => new WeatherForecast
                    {
                        Date = DateTime.Now.AddDays(index),
                        TemperatureC = rng.Next(-20, 55),
                        Summary = Summaries[rng.Next(Summaries.Length)]
                    })
                    .ToArray();
                // to show an out going dependency
                var response = await _httpClientFactory.CreateClient().GetAsync("http://google.com");
                var x = await response.Content.ReadAsStringAsync();
                scope.Span.Log("response from google: "  + x);
                scope.Span.Log("esco");
                return ret;
            }
        }
    }
}
