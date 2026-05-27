# ucantool

A tool for working with UCAN 1.0 tokens.

## Install

```sh
go install github.com/alanshaw/ucantool
```

## Usage

```sh
ucantool view <file>
```

### Examples

#### Pipe in

You can pipe invocation/delegation/container bytes in and visualize:

```sh
echo "FH4sIAAAAAAAA_1qYllySp1tm2BzJaNwU4aBS--uuQzd_z4EvGxnU2i5-zn87sTmFIfDDl6_vqh5GFLlWqH74G5ram3DxgbBGgeanz5PLLt-RYgi6uOtDwAXWhcu4FyVmeJgwvmV8yyhcWFyanJinn5lX5mCoZ6BnoFuUrGe4MjmxNKUkJTPFqjw1ySqtqDSzRC83sSg7tSQ5OTelRB-sJbG4OLWoRL8oNTk1s6AkObWiQCpLnJE_OTOxRCpLnOFjcmZxMXYzCorSGpKLS5OwyqYkFqUXL0rOLy1ZmJSfLaGVXJSYd0MrQpWBsVBIQT9v7ueGvSIW4elxHi-Ev-ssY8-6fHjTmnuvVXR4rmqGrkjNy89LTg1Q9pp9p6yn3qv6VUTe4dCmykhG96YIB1Mb5grl5uppv1xOfl-SeeuQVs2bN1le7-bqv30z99GHh0_X2Nn8OctYGN5WGPr-zfLuAMH09zc51i37fPvvnZqyhX3smKGWkpOOFGrLQaFWYQnyVHZqpVVVoHFxRrCfX0RiiHuYU1ZFbmhqYq6ld3p-mVNgUriraVZ-gGlqaGh-YmVZVlpKtpkvKGwL9MFhUaxfUFqUnJFYnIoUrjiDMz-nsTk5MScnXQ-iuSkpv6i5OcnWNlEvNbGgICcVwk7LL0rMS4dxkhLzEvMScUYDNCRNpHa7Nh-1z3Sc6Dv1rPHbOZGM15siHCo4n2XlGvs8ey0oPo8_KV1Jpe5v2_ZvqXdP8K4RXeXsYszAcfD79S3sh0x-vY36mCj9r9LFLjzrpNzF9_nzlPJPL3WChOQbxueMQjjSH4GgQEpilAR4QVFaIzxtVcp-qsz_4xh29Az7lz0WzJG3bx-e-GtRS-4E1sZbrWKJIvhT7MI0aNhDQhwavCm5qSWJS5MyUyos4K40883OzMsMKHV39XFycXcLCskqyUsN80ovLysuKMg1rvDOraqwMAqyKHINdI10M0xJyslPWpiWkpmeWlzizMjEnJKXmJuaUpJaXJJSlJ9fAnF_qJDC8QsinRwcEUcZ2CLMjkX8cF92Kv1Im5ukovaTHwp83rvOpRRnVqVKMr-AxmzGHZfeHcUPgtvf_Aovs3i_BRAAAP__8uk-f2MEAAA" | ucantool view
```

#### Token in Container

If you have a UCAN container, you can visualize a specific token by index:

```sh
ucantool view -i 1 container.ucan
```

## Screenshots

### Delegation

<img width="944" height="1056" alt="Image" src="https://github.com/user-attachments/assets/9b96ae45-79e9-4859-a5e2-37ad6eca0fa8" />

### Invocation

<img width="944" height="1056" alt="Image" src="https://github.com/user-attachments/assets/ec46f2fc-e4fe-4bfb-8927-0f2848a1d678" />

### Receipt

<img width="944" height="832" alt="Image" src="https://github.com/user-attachments/assets/5bf4db90-eee2-4d69-a3e1-2e993bf0f4eb" />

### Container

<img width="944" height="1056" alt="Image" src="https://github.com/user-attachments/assets/7f0b52b8-a112-4b90-b095-79d0ba7a07c1" />

## Contributing

Feel free to join in. All welcome. Please [open an issue](https://github.com/alanshaw/ucantool/issues)!

## License

Dual-licensed under [MIT OR Apache 2.0](LICENSE.md)
