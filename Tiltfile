tilt_dir = os.getcwd()

docker_build('ponix-all-in-one:latest',
	context=tilt_dir,
	dockerfile="./ponix-all-in-one/Dockerfile",
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
