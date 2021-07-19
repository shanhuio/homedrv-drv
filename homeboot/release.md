### Build, tag, and push homeboot image on rpi.

```
shanhu build-docker homeboot
docker tag homedrv/homeboot-images:latest \
    homedrv/homeboot-images:latest-arm64
docker push homedrv/homeboot-images:latest-arm64
```

### Do the same on an x86 computer.

```
shanhu build-docker homeboot
docker tag homedrv/homeboot-images:latest \
    homedrv/homeboot-images:latest-amd64
docker push homedrv/homeboot-images:latest-amd64
```

### Now enable experimental docker manifest command.

```
export DOCKER_CLI_EXPERIMENTAL=enabled
```

### Create and push the :latest manifest list.

```
DOCKER_CLI_EXPERIMENTAL=enabled \
    docker manifest create homedrv/homeboot:latest \
    homedrv/homeboot-images:latest-amd64 \
    homedrv/homeboot-images:latest-arm64
DOCKER_CLI_EXPERIMENTAL=enabled \
    docker manifest push -p homedrv/homeboot:latest
```
