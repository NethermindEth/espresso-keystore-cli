# Espresso Keystore CLI

This is a CLI tool that fetches secrets from Google Secret Manager and updates them with new Sequencer keys
created by the `ghcr.io/espressosystems/espresso-sequencer/sequencer:main keygen -o /keys` tool.

## How to use Docker Image

```bash
docker run --rm -it -v $(pwd)/keys:/keys -e KEYS_PATH=/keys -e PROJECT_ID=<GCP Project ID> -e SECRET_ID=<GCP Secret ID> nethermindeth/espresso-keystore-cli:v0.1.1
```

or alternatively, create a `.env` file with the following content:

```text
KEYS_PATH=${PWD}/keys
PROJECT_ID=<GCP Project ID>
SECRET_ID=<GCP Secret ID>
```

and run the following command:

```bash
docker run --rm -it --env-file .env -v $(pwd)/keys:/keys nethermindeth/espresso-keystore-cli:v0.1.1
```

### How to build Docker Image

```bash
docker build -t nethermindeth/espresso-keystore-cli:v0.1.1 .
```

## Example logs

```logs
2024/04/24 23:48:45 Secrets updated
2024/04/24 23:48:45 Sleeping for 5 seconds to allow the secret to propagate
2024/04/24 23:48:50 Final Secret Contents:
ESPRESSO_SEQUENCER_PRIVATE_STATE_KEY_0=SCHNORR_SIGNING_KEY~cHJ_u8t77fuezBa0RcEs1oW1HwqIhN0-ZH5qiawgdQVL
ESPRESSO_SEQUENCER_PRIVATE_STAKING_KEY_1=BLS_SIGNING_KEY~bU6L2o2htVBL5k75UWsH6V4BYtd4pQIXc06u4HZlegby
ESPRESSO_SEQUENCER_PRIVATE_STATE_KEY_1=SCHNORR_SIGNING_KEY~hyyljd66UmPG68U4f8Uc9j9hVVXNvynTBQHwTqDItgKu
ESPRESSO_SEQUENCER_PRIVATE_STAKING_KEY_2=BLS_SIGNING_KEY~hbt32NhY5G8-epwAyPltaW652aUJbaYCYz-IrCj3Xihs
ESPRESSO_SEQUENCER_PRIVATE_STATE_KEY_2=SCHNORR_SIGNING_KEY~UeWm87QExdba9s9ffcwH2tRJKtFoEysNcihCbs2PagGq
ESPRESSO_SEQUENCER_PRIVATE_STAKING_KEY_0=BLS_SIGNING_KEY~nZANmLcBuerhNKcSOf3nMjKlPnYzvUK2d-ZBG560zCL0
```

When there is nothing to update:

```logs
2024/04/24 23:49:08 No new secrets to update
```

## TODO:

- [ ] CI/CD pipeline to build and push docker image to registry
