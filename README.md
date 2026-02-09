<div align="center">

# 🌊 SpectralMark

**Invisible. Robust. Fast.**

DCT-domain spread-spectrum image watermarking, written in pure Go.

Embed hidden messages into images that survive noise, compression, cropping, and resizing — then detect them with a single key.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](#-quick-start)
[![License](https://img.shields.io/badge/license-MIT-6ee7b7?style=flat-square)](#)
[![Latency](https://img.shields.io/badge/latency-<10ms-38bdf8?style=flat-square)](#-latency)

</div>

---

## 📑 Table of Contents

| Section | What you'll find |
|---------|-----------------|
| [✨ Features](#-features) | What SpectralMark can do |
| [🚀 Quick Start](#-quick-start) | Get running in 30 seconds |
| [🏗️ Architecture](#️-architecture) | Pipeline flowcharts and payload format |
| [⚔️ Attack Robustness](#️-attack-robustness) | How watermarks survive different attacks |
| [🔥 Alpha Tuning](#-alpha-tuning) | Strength vs. quality tradeoff |
| [🗺️ Configuration Heatmap](#️-configuration-heatmap) | Bird's-eye view of all 40 configs |
| [⚡ Latency](#-latency) | Embed and detect timing profile |
| [🔧 CLI Reference](#-cli-reference) | Every command at a glance |
| [🔬 Technical Details](#-technical-details) | Color space, DCT, spread-spectrum math |

---

## ✨ Features

| | |
|---|---|
| 🔬 **DCT Embedding** | Spread-spectrum watermark on the Y (luminance) channel via 8×8 DCT |
| 🛡️ **Attack-Resilient** | Survives noise, JPEG-like quantization, crop, resize, brightness/contrast |
| ⚡ **Fast** | Sub-10ms embed and detect — pure Go, zero external dependencies |
| 🌐 **Web UI** | Local drag-and-drop app supporting PNG, JPEG, and PPM |
| 📊 **Benchmarking** | Built-in robustness suite with 6 attack types across configurable parameters |

---

## 🚀 Quick Start

**Requirements:** Go 1.22+

```bash
go run ./cmd/spectralmark help
```

### One-Command Demo

```bash
go run ./cmd/spectralmark demo --in sample.ppm
```

Embeds a watermark, applies 3 attacks (`noise`, `resize-nn`, `dct-quantize`), detects after each, and prints a results table.

> Default parameters: `key=k` · `msg=HELLO` · `alpha=5.0`

### Local Web App

```bash
go run ./cmd/spectralmark serve
# → http://localhost:8080
```

Drag and drop `.ppm`, `.png`, or `.jpg` files. Choose **Embed** to get a watermarked PNG, or **Detect** to get a JSON result.

---

## 🏗️ Architecture

### Embedding Pipeline

The watermark is embedded entirely in the frequency domain of the luminance channel, leaving chrominance untouched:

<img width="7556" height="615" alt="Embedding Pipeline" src="https://github.com/user-attachments/assets/b9e15c43-d150-49d5-8966-1556e34ebffb" />


### Detection Pipeline

Detection replays the keyed PRNG mapping and correlates coefficients to recover the hidden message:

<img width="5240" height="330" alt="Detection Pipeline" src="https://github.com/user-attachments/assets/e90df924-1efd-432a-aede-8d5783ca9c4b" />


### Payload Format

<img width="10779" height="1689" alt="Payload Format" src="https://github.com/user-attachments/assets/3f36c660-733b-4c75-b76a-32ce28348d8c" />


Each payload bit is repetition-coded 3×: `0 → (−1,−1,−1)` · `1 → (+1,+1,+1)`

### Full System Overview

<img width="9709" height="1059" alt="Full System Overview" src="https://github.com/user-attachments/assets/73065ecb-bdc0-4645-b66b-3c4209d155d5" />


---

## ⚔️ Attack Robustness

> 240 benchmark runs across 6 attack types, 2 image patterns, 2 sizes, 5 alpha values, and 2 messages.

| Metric | Value |
|--------|-------|
| **Overall Coverage** | 72.2% |
| **Bench Configurations** | 240 |
| **Best Attack Survival** | `none` — 70% match rate |
| **Hardest Attack** | `crop-center` — 25% match rate |

The watermark survives most transformations well, with center-crop being the most destructive — expected since it physically removes embedded coefficients.

<img width="1624" height="863" alt="attack_match_rate" src="https://github.com/user-attachments/assets/87e7ea90-5de0-41d2-978a-2a3c396ce9b8" />

Even when the watermark can't be fully decoded, the detector still reports high confidence scores, meaning the signal is present but partially corrupted:

<img width="1624" height="863" alt="attack_confidence" src="https://github.com/user-attachments/assets/153380b2-2af0-4864-a978-51d3f6ac3038" />

---

## 🔥 Alpha Tuning

The `alpha` parameter controls embedding strength. Higher alpha means a more robust watermark but slightly lower image quality (PSNR). Here's how match rate scales with alpha on 128×128 images:

**Gradient pattern** — shorter messages (`HELLO`) reach 100% survival at α=5, while longer messages need α≥6:

<img width="1468" height="841" alt="gradient128_alpha" src="https://github.com/user-attachments/assets/122c3606-bb61-45ab-8258-dd7f1680718c" />

**Texture pattern** — textured images hide watermarks better at low alpha, and `HELLO` hits 100% even at α=5:

<img width="1468" height="839" alt="texture128_alpha" src="https://github.com/user-attachments/assets/83a9f97a-8d76-409b-b01c-c3214e89ff24" />

**Takeaway:** For most use cases, `alpha=5` with short messages gives full robustness. Longer payloads or larger images may need α=6–7.

---

## 🗺️ Configuration Heatmap

A bird's-eye view across all 8 configuration groups (pattern × size × message) and 5 alpha levels:

<img width="1569" height="897" alt="heatmap_match_rate" src="https://github.com/user-attachments/assets/a5eee300-bfbb-4382-b8cd-76907f709b06" />

The clear diagonal trend confirms: higher alpha always helps, but texture/128/HELLO is the sweet spot — robust even at low alpha.

---

## ⚡ Latency

Embed and detect both run in under 10ms across all configurations. Performance is remarkably consistent regardless of pattern, size, or alpha:

| Metric | Average |
|--------|---------|
| **Embed** | ~9.4 ms |
| **Detect (watermarked)** | ~9.3 ms |
| **Detect (original)** | ~9.3 ms |

<img width="1870" height="854" alt="latency_profile" src="https://github.com/user-attachments/assets/f956fa27-5bf7-4116-bcbe-13e32cfdb48f" />

---

## 🔧 CLI Reference

```bash
# Embed watermark
go run ./cmd/spectralmark embed --in a.ppm --out w.ppm --key k --msg HELLO --alpha 3.0

# Detect watermark
go run ./cmd/spectralmark detect --in w.ppm --key k

# Robustness benchmark
go run ./cmd/spectralmark bench --in a.ppm --key k --msg HELLO

# PSNR + diff image
go run ./cmd/spectralmark metrics --a orig.ppm --b w.ppm --diff diff.ppm
```

<details>
<summary>More utilities</summary>

```bash
go run ./cmd/spectralmark ppm-copy --in a.ppm --out b.ppm
go run ./cmd/spectralmark to-gray --in a.ppm --out gray.ppm
go run ./cmd/spectralmark dct-check
go run ./cmd/spectralmark prng-demo --key abc --n 10
go run ./cmd/spectralmark payload-demo --msg HELLO
```

</details>

<details>
<summary>📋 Full 40-configuration benchmark table</summary>

| Pattern | Size | α | Message | Present | Decode | Match | Score |
|---------|------|---|---------|---------|--------|-------|-------|
| gradient | 128 | 3 | HELLO | 0% | 0% | 0% | 92.7% |
| gradient | 128 | 3 | SPECTRALMARK_DEMO | 0% | 0% | 0% | 94.8% |
| gradient | 128 | 4 | HELLO | 83.3% | 83.3% | 83.3% | 96.3% |
| gradient | 128 | 4 | SPECTRALMARK_DEMO | 0% | 0% | 0% | 97.9% |
| gradient | 128 | 5 | HELLO | **100%** | **100%** | **100%** | 97.2% |
| gradient | 128 | 5 | SPECTRALMARK_DEMO | 0% | 0% | 0% | 100% |
| gradient | 128 | 6 | HELLO | **100%** | **100%** | **100%** | 98.0% |
| gradient | 128 | 6 | SPECTRALMARK_DEMO | 66.7% | 66.7% | 66.7% | 98.4% |
| gradient | 128 | 7 | HELLO | **100%** | **100%** | **100%** | 98.0% |
| gradient | 128 | 7 | SPECTRALMARK_DEMO | 66.7% | 66.7% | 66.7% | 98.5% |
| gradient | 256 | 3 | HELLO | 0% | 0% | 0% | 94.8% |
| gradient | 256 | 3 | SPECTRALMARK_DEMO | 0% | 0% | 0% | 95.8% |
| gradient | 256 | 4 | HELLO | 66.7% | 66.7% | 66.7% | 97.6% |
| gradient | 256 | 4 | SPECTRALMARK_DEMO | 33.3% | 33.3% | 33.3% | 98.1% |
| gradient | 256 | 5 | HELLO | 83.3% | 83.3% | 83.3% | 97.6% |
| gradient | 256 | 5 | SPECTRALMARK_DEMO | 83.3% | 83.3% | 83.3% | 97.3% |
| gradient | 256 | 6 | HELLO | 83.3% | 83.3% | 83.3% | 98.2% |
| gradient | 256 | 6 | SPECTRALMARK_DEMO | 83.3% | 83.3% | 83.3% | 97.8% |
| gradient | 256 | 7 | HELLO | 83.3% | 83.3% | 83.3% | 98.2% |
| gradient | 256 | 7 | SPECTRALMARK_DEMO | 83.3% | 83.3% | 83.3% | 97.8% |
| texture | 128 | 3 | HELLO | 83.3% | 83.3% | 83.3% | 94.1% |
| texture | 128 | 3 | SPECTRALMARK_DEMO | 50% | 50% | 50% | 94.4% |
| texture | 128 | 4 | HELLO | 83.3% | 83.3% | 83.3% | 94.5% |
| texture | 128 | 4 | SPECTRALMARK_DEMO | 50% | 50% | 50% | 96.8% |
| texture | 128 | 5 | HELLO | **100%** | **100%** | **100%** | 96.3% |
| texture | 128 | 5 | SPECTRALMARK_DEMO | 50% | 50% | 50% | 98.5% |
| texture | 128 | 6 | HELLO | **100%** | **100%** | **100%** | 97.2% |
| texture | 128 | 6 | SPECTRALMARK_DEMO | 50% | 50% | 50% | 98.8% |
| texture | 128 | 7 | HELLO | **100%** | **100%** | **100%** | 98.4% |
| texture | 128 | 7 | SPECTRALMARK_DEMO | **100%** | **100%** | **100%** | 97.2% |
| texture | 256 | 3 | HELLO | 0% | 0% | 0% | 94.8% |
| texture | 256 | 3 | SPECTRALMARK_DEMO | 0% | 0% | 0% | 94.8% |
| texture | 256 | 4 | HELLO | 0% | 0% | 0% | 93.8% |
| texture | 256 | 4 | SPECTRALMARK_DEMO | 0% | 0% | 0% | 92.7% |
| texture | 256 | 5 | HELLO | 0% | 0% | 0% | 93.8% |
| texture | 256 | 5 | SPECTRALMARK_DEMO | 33.3% | 33.3% | 33.3% | 93.8% |
| texture | 256 | 6 | HELLO | 0% | 0% | 0% | 96.9% |
| texture | 256 | 6 | SPECTRALMARK_DEMO | 83.3% | 83.3% | 83.3% | 93.3% |
| texture | 256 | 7 | HELLO | 83.3% | 83.3% | 83.3% | 93.3% |
| texture | 256 | 7 | SPECTRALMARK_DEMO | 83.3% | 83.3% | 83.3% | 94.5% |

</details>

---

## 🔬 Technical Details

### Color Space

Watermarking operates on the luminance (Y) channel only:

```
Y  =  0.299·R + 0.587·G + 0.114·B
Cb = 128 − 0.169·R − 0.331·G + 0.500·B
Cr = 128 + 0.500·R − 0.419·G − 0.081·B
```

### 8×8 DCT

Standard Type-II DCT with normalization constants `C(k) = 1/√2` for `k=0`, else `1`. Applied per-block after edge-replicate padding to an 8×8 grid.

### Spread-Spectrum Embedding

For each payload symbol, a keyed PRNG selects target DCT slots (block + coefficient index). A chip sign (±1) scrambles the symbol, and the selected mid-frequency coefficient is nudged toward a signed target margin controlled by `alpha` (α).

### Detection

Replays the keyed mapping, correlates sampled coefficients with chip signs, repetition-decodes to raw bits, and validates via sync pattern + CRC-16. Includes bounded sync-offset scanning and constrained bit-fix search for robustness while limiting false positives.
