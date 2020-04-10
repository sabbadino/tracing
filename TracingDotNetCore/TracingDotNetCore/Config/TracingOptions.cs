using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace TracingDotNetCore.Config
{
    public class TracingOptions
    {
        public string TracerTarget { get; set; } = "Jaeger"; //, // "datadog"
        public bool EnableOpenTracingAutoTracing { get; set; } = true;
        public string JaegerHost { get; set; } 
        public string DataDogAgentUrl { get; set; }
        public bool EnableDataDogEnableAutoTracing { get; set; } = false;
        public int JaegerPort { get; set; }
    }
}
