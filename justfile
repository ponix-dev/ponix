basedir := `pwd`
artifacts_dir := basedir / "artifacts"

default:
    echo {{artifacts_dir}}

board:
    lsusb | grep UART

int_probe_rs:
    #!/usr/bin/env bash
    curl --proto '=https' --tlsv1.2 -LsSf https://github.com/probe-rs/probe-rs/releases/latest/download/probe-rs-tools-installer.sh | sh
    sudo usermod -a -G dialout $USER
    sudo cp ./utils/69-probe-rs.rules /etc/udev/rules.d
    sudo udevadm control --reload
    sudo udevadm trigger

init_esp:
    #!/usr/bin/env bash
    cargo install espup
    espup install
    . $HOME/export-esp.sh
    sudo apt install -y pkg-config libusb-1.0-0-dev libftdi1-dev
    sudo apt-get -y install libudev-dev
    cargo install esp-generate
    cargo install espflash

init: int_probe_rs init_esp

dev-up:
    #!/usr/bin/env bash
    . $HOME/export-esp.sh

otel-up:
    docker compose -f otel.docker-compose.yaml up -d

otel-down:
    docker compose -f otel.docker-compose.yaml down

