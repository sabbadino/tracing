#See https://aka.ms/containerfastmode to understand how Visual Studio uses this Dockerfile to build your images for faster debugging.

FROM mcr.microsoft.com/dotnet/core/aspnet:3.1.3 AS base
WORKDIR /app
EXPOSE 80

FROM mcr.microsoft.com/dotnet/core/sdk:3.1.201 AS build
WORKDIR /src
COPY ["TracingDotNetCore/TracingDotNetCore.csproj", "TracingDotNetCore/"]
RUN dotnet restore "TracingDotNetCore/TracingDotNetCore.csproj"
COPY . .
WORKDIR "/src/TracingDotNetCore"
RUN dotnet build "TracingDotNetCore.csproj" -c Release -o /app/build

FROM build AS publish
RUN dotnet publish "TracingDotNetCore.csproj" -c Release -o /app/publish

FROM base AS final
WORKDIR /
RUN curl -LO https://github.com/DataDog/dd-trace-dotnet/releases/download/v1.16.0/datadog-dotnet-apm_1.16.0_amd64.deb
RUN dpkg -i ./datadog-dotnet-apm_1.16.0_amd64.deb
# RUN ls /
# RUN ls /opt
# RUN ls /opt/datadog
# DATADOG
ENV CORECLR_PROFILER={846F5F1C-F9AE-4B07-969E-05C26BC060D8}
ENV CORECLR_PROFILER_PATH=/opt/datadog/Datadog.Trace.ClrProfiler.Native.so
ENV DD_INTEGRATIONS=/opt/datadog/integrations.json

# can be setup at application startup or in env variables in config map : 
#ENV CORECLR_ENABLE_PROFILING=1
#ENV DD_TRACE_ANALYTICS_ENABLED=true
#ENV DD_LOGS_INJECTION=true
#ENV DD_ENV=
#ENV DD_AGENT_HOST=dmbo-datadog-agent.westeurope.cloudapp.azure.com
#ENV DD_SERVICE_NAME=dmbo.v3.auditing.api

RUN /opt/datadog/createLogPath.sh

WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "TracingDotNetCore.dll"]