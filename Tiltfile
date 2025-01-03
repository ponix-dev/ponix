tilt_dir = os.getcwd()

docker_build('management-service:latest',
	context=tilt_dir,
	dockerfile="./cmd/management-service/Dockerfile",
    only=[
        "./cmd",
        "./internal",
        "./go.mod",
        "./go.sum",
        "./config",
        "./.mise.toml",
        "./gen"
    ],
	ignore=[
		"**/*_test.go",
	]
)


docker_compose('./docker-compose.yaml')
