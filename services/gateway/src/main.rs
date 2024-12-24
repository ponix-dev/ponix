use rand::Rng;
use tokio::time;

#[tokio::main]
async fn main() {
    let resource = telemetry::providers::init_resource(String::from("dice_roll_service"));
    telemetry::providers::init_meter_provider(&resource);

    let gauge = telemetry::metrics::guage(
        String::from("dice_roll"),
        String::from("the value of a dice roll"),
        String::from("side"),
    );

    loop {
        let random_number = rand::thread_rng().gen_range(1..7);
        println!("{}", random_number);

        gauge.record(f64::from(random_number), &[]);
        time::sleep(time::Duration::new(3, 0)).await
    }
}
