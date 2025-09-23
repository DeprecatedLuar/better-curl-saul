<h3 align="center">When HTTP gets complicated...</h3>
<p align="center">
  <img src="other/assets/saul-logo (1).png" width="600"/>
</p>

<p align="center">
  <a href="https://github.com/DeprecatedLuar/better-curl-saul/stargazers">
    <img src="https://img.shields.io/github/stars/DeprecatedLuar/better-curl-saul?style=for-the-badge&logo=github&color=1f6feb&logoColor=white&labelColor=black"/>
  </a>
  <a href="https://github.com/DeprecatedLuar/better-curl-saul/releases">
    <img src="https://img.shields.io/github/v/release/DeprecatedLuar/better-curl-saul?style=for-the-badge&logo=go&color=00ADD8&logoColor=white&labelColor=black"/>
  </a>
  <a href="https://github.com/DeprecatedLuar/better-curl-saul/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/DeprecatedLuar/better-curl-saul?style=for-the-badge&color=green&labelColor=black"/>
  </a>
</p>

---



<p align="center">
  <img src="other/assets/saul-catboy-final.png" width="700"/>
</p>

## Live Demo

<p align="center">
  <img src="other/assets/demo.gif" alt="Better-Curl Demo" width="800"/>
</p>

## The Problem

**In a nutshell, this is disgusting:**
```bash
curl -X POST https://api.github.com/repos/owner/repo/issues \
  -H "Authorization: Bearer ghp_token123" \
  -H "Content-Type: application/json" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{
    "title": "Bug Report",
    "body": "Something is broken",
    "labels": ["bug", "priority-high"],
    "assignees": ["developer1", "developer2"]
  }'
```

## **Try this instead:**

```bash
saul github set url https://api.github.com/repos/owner/repo/issues
saul set method POST
saul set header Authorization="Bearer {@token}"
saul set body title="Bug Report" body="Something is broken" labels=[bug,priority-high] assignees=[developer1,developer2]
saul call
```

## The cool stuff you've never seen before

- **Workspace-based** - Each API gets its own organized folder
- **Smart variables** - `{@token}` persists,`{?name}` prompts every time
- **Response filtering** - Show only the fields you care about
- **Git-friendly** - TOML files version control beautifully
- **Unix composable** - Script it, pipe it, shell it

<img src="other/assets/saul-hd-wide.png" width="800"/>

## 📦 Installation

**Supports:** Linux, macOS, Windows (all architectures)

### One-Line Install (Easiest)
```bash
curl -sSL https://raw.githubusercontent.com/DeprecatedLuar/better-curl-saul/releases/install.sh | bash
```

### Manual Install
1. Download binary for your OS from [releases](https://github.com/DeprecatedLuar/better-curl-saul/releases)
2. Make executable: `chmod +x saul-*`
3. Move to PATH: `sudo mv saul-* /usr/local/bin/saul`

### From Source (Try-Harders)
```bash
git clone https://github.com/DeprecatedLuar/better-curl-saul.git
cd better-curl-saul
./other/install-local.sh  # Local development build
```

>[!NOTE]
> One-line install automatically downloads pre-built binaries or builds from source as fallback

<h1 align=center> Quick Start </h1 align=center>



```bash
# Create a test workspace
saul demo set url https://jsonplaceholder.typicode.com/posts/1
saul demo set method GET
saul demo call

# Try with variables
saul api set url https://httpbin.org/post
saul api set method POST
saul api set body name={?your_name} message="Hello from Saul"
saul api call

# Oh... yeah, for nesting just use dot notation like obj.field=idk
```

## 📖 Core Commands

```bash
# Configure your API workspace (or preset, same thing)
saul [workspace] set url https://api.example.com
saul set method POST
saul set header Authorization="Bearer {@token}"
saul set body user.name={?username} user.email=john@test.com

# Execute the request
saul call

# Check your configuration, note that preste/workspace name keeps
# stored in memory after first mention on syntax
saul [workspace] check url
saul [workspace] check body

# View response history
saul [workspace] check history
```

## 🗺️ Roadmap

- [x] Workspace-based configuration
- [x] Smart variable system (`{@}` / `{?}`)
- [x] Response filtering
- [x] Response history
- [x] Terminal session memory
- [x] Bulk operations
- [ ] User configuration system
- [ ] Actual Documentation
- [ ] Think of more features

## Little Note

**Beta software** - Core features work, documentation in progress.

Bug or feedback? I will be very, very, very happy if you let me know your thoughts.

<img src="other/assets/saul-pointing.png" width="800"/>

---

<p align="center">
  <a href="https://github.com/DeprecatedLuar/better-curl-saul/issues">
    <img src="https://img.shields.io/badge/Found%20a%20bug%3F-Report%20it!-red?style=for-the-badge&logo=github&logoColor=white&labelColor=black"/>
  </a>
</p>

