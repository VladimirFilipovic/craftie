## Usage examples

# Start a 2-hour session

./craftie start -p "my-project" -e 2h

# Start a 30-minute session

./craftie start -p "my-project" -e 30m

# Start without end time

./craftie start -p "my-project"

## Shell Hook

Craftie can notify you when you `cd` into a project directory, asking if you want to start a session.

### Setup

1. Add your project directories to `~/.config/craftie/craftie.yaml`:

```yaml
notifications:
  enabled: true
  projects:
    myproject: "~/Documents/myproject"
    webapp: "~/projects/webapp"
```

2. Add the hook to your shell config by appending the output of the setup command:

**zsh**:

```zsh
craftie hook setup --s zsh >> ~/.zshrc
```

**bash**:

```bash
craftie hook setup --s bash >> ~/.bashrc
```

3. Restart your shell or source the config file.

When you `cd` into a configured project directory, a desktop notification will ask if you want to start tracking. Selecting "Start Session" pre-fills the start command in your terminal.
