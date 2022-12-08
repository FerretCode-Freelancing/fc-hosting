# Welcome to Tilt!
#   To get you started as quickly as possible, we have created a
#   starter Tiltfile for you.
#
#   Uncomment, modify, and delete any commands as needed for your
#   project's configuration.


# Output diagnostic messages
#   You can print log messages, warnings, and fatal errors, which will
#   appear in the (Tiltfile) resource in the web UI. Tiltfiles support
#   multiline strings and common string operations such as formatting.
#
#   More info: https://docs.tilt.dev/api.html#api.warn

# Build Docker image
#   Tilt will automatically associate image builds with the resource(s)
#   that reference them (e.g. via Kubernetes or Docker Compose YAML).
#
#   More info: https://docs.tilt.dev/api.html#api.docker_build_with_restart
#
# docker_build('registry.example.com/my-image',
#              context='.',
#              # (Optional) Use a custom Dockerfile path
#              dockerfile='./deploy/app.dockerfile',
#              # (Optional) Filter the paths used in the build
#              only=['./app'],
#              # (Recommended) Updating a running container in-place
#              # https://docs.tilt.dev/live_update_reference.html
#              live_update=[
#                 # Sync files from host to container
#                 sync('./app', '/src/'),
#                 # Execute commands inside the container when certain
#                 # paths change
#                 run('/src/codegen.sh', trigger=['./app/api'])
#              ]
# )
#

load('ext://restart_process', 'docker_build_with_restart')

docker_build_with_restart('sthanguy/fc-auth',
							context='./services/auth',
							entrypoint='node .',
							dockerfile='./services/auth/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/auth', '/usr/route'),
							]
)

docker_build_with_restart('sthanguy/fc-gateway',
							context='./services/gateway',
							entrypoint='go run main.go',
							dockerfile='./services/gateway/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/gateway', '/usr/gateway'),
							]
)

docker_build_with_restart('sthanguy/fc-subscribe',
							context='./services/subscribe',
							entrypoint='node .',
							dockerfile='./services/subscribe/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/subscribe', '/usr/route'),
							]
)

docker_build_with_restart('sthanguy/fc-upload',
							context='./services/upload',
							entrypoint='go run main.go',
							dockerfile='./services/upload/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/upload', '/usr/route'),
								run('cd /usr/route && go build -v -o /usr/route'),
								run('route')
							]
)

docker_build_with_restart('sthanguy/fc-projects',
							context='./services/projects',
							entrypoint='go run main.go',
							dockerfile='./services/projects/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/projects', '/usr/route'),
								run('cd /usr/route && go build -v -o /usr/route'),
								run('route')
							]
)

docker_build_with_restart('sthanguy/fc-provision',
							context='../fc-provision',
							dockerfile='../fc-provision/Dockerfile',
							entrypoint='builder',
							extra_tag='sthanguy/fc-provision:latest',
							live_update=[
								sync('../fc-provision', '/usr/src/provision'),
								run('cd /usr/src/provision && go build -v -o /usr/local/bin/builder'),
								run('builder')
							]
)

docker_build_with_restart('sthanguy/fc-deploy',
							context='./services/deploy',
							dockerfile='./services/deploy/Dockerfile',
							entrypoint='go run main.go',
							extra_tag='latest',
							live_update=[
								sync('./services/deploy', '/usr/route'),
								run('cd /usr/route && go build -v -o /usr/route'),
								run('route')
							]
)

# Apply Kubernetes manifests
#   Tilt will build & push any necessary images, re-deploying your
#   resources as they change.
#
#   More info: https://docs.tilt.dev/api.html#api.k8s_yaml
#
# k8s_yaml(['k8s/deployment.yaml', 'k8s/service.yaml'])

# deployments
k8s_yaml(['kubernetes/deployments/auth.yml', 'kubernetes/deployments/gateway.yml', 'kubernetes/deployments/subscribe.yml', 'kubernetes/deployments/upload.yml', 'kubernetes/deployments/cache.yml', 'kubernetes/deployments/registry.yml', 'kubernetes/deployments/provision.yml', 'kubernetes/deployments/projects.yml', 'kubernetes/deployments/deploy.yml'])
# services
k8s_yaml(['kubernetes/services/auth.yml', 'kubernetes/services/gateway.yml', 'kubernetes/services/subscribe.yml', 'kubernetes/services/upload.yml', 'kubernetes/services/cache.yml', 'kubernetes/services/registry.yml', 'kubernetes/services/provision.yml', 'kubernetes/services/projects.yml', 'kubernetes/services/deploy.yml'])
# pvcs
k8s_yaml(['kubernetes/pvc/registry.yml'])
# ingress
k8s_yaml(['kubernetes/ingress/gateway.yml'])

# Customize a Kubernetes resource
#   By default, Kubernetes resource names are automatically assigned
#   based on objects in the YAML manifests, e.g. Deployment name.
#
#   Tilt strives for sane defaults, so calling k8s_resource is
#   optional, and you only need to pass the arguments you want to
#   override.
#
#   More info: https://docs.tilt.dev/api.html#api.k8s_resource
#
# k8s_resource('my-deployment',
#              # map one or more local ports to ports on your Pod
#              port_forwards=['5000:8080'],
#              # change whether the resource is started by default
#              auto_init=False,
#              # control whether the resource automatically updates
#              trigger_mode=TRIGGER_MODE_MANUAL
# )


# Run local commands
#   Local commands can be helpful for one-time tasks like installing
#   project prerequisites. They can also manage long-lived processes
#   for non-containerized services or dependencies.
#
#   More info: https://docs.tilt.dev/local_resource.html
#
# local_resource('install-helm',
#                cmd='which helm > /dev/null || brew install helm',
#                # `cmd_bat`, when present, is used instead of `cmd` on Windows.
#                cmd_bat=[
#                    'powershell.exe',
#                    '-Noninteractive',
#                    '-Command',
#                    '& {if (!(Get-Command helm -ErrorAction SilentlyContinue)) {scoop install helm}}'
#                ]
# )


# Extensions are open-source, pre-packaged functions that extend Tilt
#
#   More info: https://github.com/tilt-dev/tilt-extensions
#
load('ext://git_resource', 'git_checkout')


# Organize logic into functions
#   Tiltfiles are written in Starlark, a Python-inspired language, so
#   you can use functions, conditionals, loops, and more.
#
#   More info: https://docs.tilt.dev/tiltfile_concepts.html
#
def tilt_demo():
    # Tilt provides many useful portable built-ins
    # https://docs.tilt.dev/api.html#modules.os.path.exists
    if os.path.exists('tilt-avatars/Tiltfile'):
        # It's possible to load other Tiltfiles to further organize
        # your logic in large projects
        # https://docs.tilt.dev/multiple_repos.html
        load_dynamic('tilt-avatars/Tiltfile')
    watch_file('tilt-avatars/Tiltfile')
    git_checkout('https://github.com/tilt-dev/tilt-avatars.git',
                 checkout_dir='tilt-avatars')


# Edit your Tiltfile without restarting Tilt
#   While running `tilt up`, Tilt watches the Tiltfile on disk and
#   automatically re-evaluates it on change.
#
#   To see it in action, try uncommenting the following line with
#   Tilt running.
# tilt_demo()
