use opentelemetry::{global, KeyValue};
use opentelemetry_otlp::MetricExporter;
use opentelemetry_sdk::metrics::{PeriodicReader, SdkMeterProvider};
use opentelemetry_sdk::{runtime, Resource};
use std::time::Duration;

pub fn init_resource(service_name: String) -> Resource {
    Resource::new([KeyValue::new("service.name", service_name)])
}

pub fn init_meter_provider(resource: &Resource) {
    let exporter = MetricExporter::builder().with_tonic().build().unwrap();

    let reader = PeriodicReader::builder(exporter, runtime::Tokio)
        .with_interval(Duration::new(1, 0))
        .build();

    let provider = SdkMeterProvider::builder()
        .with_reader(reader)
        .with_resource(resource.clone())
        .build();

    global::set_meter_provider(provider.clone());
}
