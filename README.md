# DigitalOcean-to-Porkbun DNS Migrator

Copy DNS records from DigitalOcean to Porkbun.

This is a very simple tool intended for one-time migrations with a human involved in the process; it is not sufficiently robust to work without supervision or to be run repeatedly.

In particular, this tool is _not_ idempotent:

- It will attempt to copy every record, every time it is run, without regard for whether a similar record already exists in the destination.
- In the event of a failure partway through a migration, this tool will not resume from where it left off; it will start over from the beginning. 

This tool will not edit or remove existing records in Porkbun under any circumstances.

## Build

Because this is a one-off tool for occasional use, I'm not providing a CI pipeline or prebuilt binaries/Docker images. To get a binary, check out the repo and build it:

```
git clone https://github.com/cdzombak/dns-do-to-porkbun.git
cd dns-do-to-porkbun
go build -o out .
```

The only requirement is a working Go installation.

## Usage

```
DO_API_TOKEN=dop_v1_foobarbaz \
PB_API_SECRET=sk1_01234567 \
PB_API_KEY=pk1_abcdabcd \
./out/dns-do-to-porkbun -domain MYDOMAIN.COM [-dry-run=false] 
```

- Flag `-domain`: The domain name for DNS migration.
- Flag `-dry-run`: If true, the tool will not make any changes to Porkbun; it will only print what it would do. `True` by default; you must pass `-dry-run=false` to make any changes to Porkbun.
- Environment variable `DO_API_TOKEN`: DigitalOcean API token with read access.
- Environment variable `PB_API_SECRET`: Porkbun API secret key.
- Environment variable `PB_API_KEY`: Porkbun API key.

> [!NOTE]  
> On the Porkbun side, you need to enable API access for each domain you want to migrate. See [Porkbun's API docs](https://kb.porkbun.com/article/190-getting-started-with-the-porkbun-api).

## Configuration

Environment variables may be placed in the special file `.env`. This file will be read automatically from the working directory if it exists. See [`env.sample`](env.sample) for an example.

## API Tokens/Management

- DigitalOcean: https://cloud.digitalocean.com/account/api/tokens
- Porkbun: https://porkbun.com/account/api

## License

LGPL 3.0; see [LICENSE](LICENSE) in this repository.

## Author

Chris Dzombak ([dzombak.com](https://www.dzombak.com); [GitHub @cdzombak](https://github.com/cdzombak))
