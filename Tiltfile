load('ext://restart_process', 'docker_build_with_restart')

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/version-metrics ./'

local_resource(
    'version-metrics-compile',
    compile_cmd,
    deps=['./main.go'],
)

docker_build_with_restart(
    'version-metrics-tilt',
    '.',
    entrypoint=['/app/build/version-metrics'],
    dockerfile='Dockerfile.tilt',
    only=[
        './build',
    ],
    live_update=[
        sync('./build', '/app/build'),
    ],
)

#docker_build('version-metrics', '.', dockerfile='Dockerfile')
k8s_yaml('k8s.yaml')
k8s_resource('version-metrics', port_forwards=8000,
             resource_deps=['version-metrics-compile'])
