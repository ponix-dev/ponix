use opentelemetry::global;

pub fn guage(
    name: String,
    description: String,
    unit: String,
) -> opentelemetry::metrics::Gauge<f64> {
    //TODO: fix meter name to something configurable
    let meter = global::meter("ponix-gateway");

    let gauge = meter
        .f64_gauge(name)
        .with_description(description)
        .with_unit(unit)
        .build();

    gauge
}
