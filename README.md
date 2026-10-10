# Homelab Diagram Generator

[![ACR B4c](https://img.shields.io/badge/ACR-B4c-2140B5)](ACR.md)

Describe your homelab in a YAML file and get a clean, shareable SVG diagram of it.

`hdg` is a small command-line tool written in Go. You list your devices, VLANs, and services in one YAML file, and it lays them out as cards grouped into sections. You don't drag boxes around in a drawing app, and when your homelab changes you edit a line of YAML and regenerate.

![Sample homelab diagram](https://i.imgur.com/bZ5o53D.png)

## Features

- **Diagrams as code:** your diagram is a text file you can keep in git, diff, and review.
- **VLAN colour coding:** each VLAN gets a colour, shown in the legend and as a stripe on every device card in that VLAN.
- **Sections and groups:** organise devices into sections (Network Core, Servers, IoT, and so on), and cluster related devices into labelled groups, such as the VMs and containers on a Proxmox host.
- **Group defaults:** set shared values like icon, OS, or VLAN once per group instead of on every device.
- **Badges:** mark devices with small lettered circles for things like "Managed by Ansible", "Runs Docker", or "Web-Facing".
- **Services:** list the apps running on each device inside its card.
- **23 built-in icons:** routers, switches, servers, NAS boxes, Raspberry Pis, VMs, containers, TVs, IoT devices, and more.
- **No dependencies at runtime:** a single static binary with the icons embedded. The output is a standalone SVG.

## Installation

### Pre-built binary (Linux x86-64)

The repository includes a pre-built `hdg` binary for 64-bit Linux:

```sh
git clone https://github.com/gregdotca/homelab-diagram-generator.git
cd homelab-diagram-generator
./hdg -input sample.yaml -output sample.svg
```

### Build from source

You need Go 1.21 or newer.

```sh
git clone https://github.com/gregdotca/homelab-diagram-generator.git
cd homelab-diagram-generator
./build.sh
```

This builds `./hdg` for your platform.

### go install

```sh
go install github.com/gregdotca/homelab-diagram-generator@latest
```

Go names the installed binary after the repository, so this gives you `homelab-diagram-generator` rather than `hdg`.

## Quick start

1. Copy `sample.yaml` to `homelab.yaml`.
2. Edit `homelab.yaml` to describe your homelab.
3. Run `./hdg` to generate `homelab.svg`.

`homelab.yaml` and `homelab.svg` are in `.gitignore`, so your own homelab's details won't be committed by accident.

## Usage

```sh
./hdg [-input FILE] [-output FILE]
```

| Flag | Default | Description |
|---|---|---|
| `-input` | `homelab.yaml` | The YAML file describing your homelab |
| `-output` | `homelab.svg` | Where to write the SVG diagram |

## Configuration

A config file has seven top-level keys, all optional, though a useful diagram needs at least `sections` and `devices`. [`sample.yaml`](sample.yaml) uses every feature and has comments explaining each one, so it's the best place to start.

### `title`

The heading at the top of the diagram. Defaults to "Homelab Diagram".

### `vlans`

Your network segments. Colours are assigned automatically if you leave them out.

```yaml
vlans:
  - { id: servers, label: "VLAN 10 - Servers - 10.0.10.0/24", color: "#4CAF50" }
  - { id: iot, label: "VLAN 20 - IoT - 10.0.20.0/24" }
```

### `badges`

Small coloured circles in the corner of device cards. The letter defaults to the first letter of the `id`, and colours are assigned automatically.

```yaml
badges:
  - { id: docker, label: "Runs Docker" }
  - { id: web, label: "Web-Facing", letter: "W" }
```

### `sections`

The large boxes the diagram is divided into. They're drawn top to bottom in the order you list them.

```yaml
sections:
  - { id: network, label: "Network Core" }
  - { id: servers, label: "Infrastructure" }
```

### `devices`

Standalone devices. Each becomes a card in its section.

```yaml
devices:
  - id: nas
    label: "NAS"
    section: servers
    icon: nas
    hardware: "Synology DS920+"
    os: "DSM 7"
    specs: "16TB usable (RAID), Ethernet"
    ip: "10.0.10.5"
    vlan: servers
    badges: [docker]
    services: ["Plex", "Time Machine"]
```

| Field | Required | Description |
|---|---|---|
| `id` | Yes | Unique identifier |
| `label` | Yes | Display name |
| `section` | Yes | The section this device belongs in |
| `icon` | No | One of the built-in icons below. Defaults to `generic`. |
| `hardware` | No | Hardware description line |
| `os` | No | Operating system line |
| `specs` | No | An extra details line, shown last. Good for the connection type. |
| `ip` | No | IP address or hostname, shown at the bottom of the card |
| `vlan` | No | A VLAN `id`. Adds that VLAN's colour stripe to the card. |
| `badges` | No | A list of badge `id`s |
| `services` | No | A list of apps or services shown inside the card |

Devices also have a compact one-line form:

```yaml
  - { id: phone, label: "Phone", section: clients, icon: phone, specs: "WiFi", vlan: trusted }
```

### `groups`

Groups cluster related devices inside a section, like the VMs on a hypervisor. When a section has more than one group, each group gets its own labelled box.

```yaml
groups:
  - id: vms
    label: "Virtual Machines"
    section: virtualization
    columns: 3
    defaults:
      icon: vm
      vlan: servers
      badges: [monitored]
    devices:
      - { id: vm-docker, label: "Docker Host", os: "Debian 12", ip: "10.0.10.20" }
      - { id: vm-hass, label: "Home Assistant", os: "HAOS 12", ip: "10.0.10.22" }
```

- `columns` sets how many cards go in each row. It's worked out automatically if you leave it out.
- `defaults` can set `icon`, `os`, `hardware`, `vlan`, and `badges` for every device in the group. A device's own value overrides the default.
- `style` sets the group box's `background` and `border` colours.

### `style`

Optional overrides for the look of the whole diagram. Every value below is the default, except `font_family`, which defaults to Inter with common system fonts as fallbacks:

```yaml
style:
  background: "#ffffff"
  font_family: "Inter, sans-serif"
  card_width: 220
  card_padding: 8
  group_padding: 10
  group_spacing: 15
```

## Icons

`cloud`, `container`, `desktop`, `game-console`, `generic`, `iot`, `kindle`, `laptop`, `modem`, `nas`, `phone`, `printer`, `raspberry-pi`, `router`, `scale`, `server`, `switch`, `tablet`, `tv`, `vacuum`, `vm`, `watch`, and `wifi-ap`.

## An actual homelab

A real homelab, generated with `hdg`:

![A real homelab diagram](https://i.imgur.com/T6Ed2ck.png)

## License

Homelab Diagram Generator is released under the [MIT License](LICENSE). You're free to use, modify, and share it, including commercially, as long as you keep the copyright notice.
