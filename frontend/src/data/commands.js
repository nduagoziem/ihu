// Curated Linux command reference, grouped by toolchain.
// Each command has a short meaning suitable for quick lookup.

export const commandGroups = [
  {
    id: 'native',
    name: 'Native Linux',
    color: '#3b82f6',
    commands: [
      { cmd: 'ls -la', meaning: 'List all files including hidden, with details' },
      { cmd: 'cd <path>', meaning: 'Change current directory' },
      { cmd: 'pwd', meaning: 'Print working directory path' },
      { cmd: 'cp -r src dst', meaning: 'Recursively copy files/directories' },
      { cmd: 'mv old new', meaning: 'Move or rename a file/directory' },
      { cmd: 'rm -rf <path>', meaning: 'Force-remove files recursively (careful)' },
      { cmd: 'mkdir -p a/b/c', meaning: 'Create nested directories' },
      { cmd: 'touch <file>', meaning: 'Create empty file or update timestamp' },
      { cmd: 'chmod 755 <file>', meaning: 'Set file permissions (rwxr-xr-x)' },
      { cmd: 'chown user:group <path>', meaning: 'Change file ownership' },
      { cmd: 'sudo -i -u <user>', meaning: 'Start interactive shell as another user' },
      { cmd: 'sudo <cmd>', meaning: 'Run a single command as root' },
      { cmd: 'apt update && apt upgrade', meaning: 'Refresh package index then upgrade' },
      { cmd: 'cat <file>', meaning: 'Print file contents to stdout' },
      { cmd: 'grep -rn "term" .', meaning: 'Recursive line search with line numbers' },
      { cmd: 'find . -name "*.go"', meaning: 'Find files matching a pattern' },
      { cmd: 'tar -xzf archive.tar.gz', meaning: 'Extract gzip tarball' },
      { cmd: 'df -h', meaning: 'Disk usage in human-readable form' },
      { cmd: 'free -h', meaning: 'Memory usage in human-readable form' },
      { cmd: 'top / htop', meaning: 'Live process and resource monitor' },
      { cmd: 'ps aux', meaning: 'Snapshot of all running processes' },
      { cmd: 'kill -9 <pid>', meaning: 'Force-terminate a process by ID' },
      { cmd: 'curl -O <url>', meaning: 'Download a file preserving its name' },
      { cmd: 'wget <url>', meaning: 'Non-interactive file downloader' },
      { cmd: 'ssh user@host', meaning: 'Open a remote secure shell session' },
      { cmd: 'scp file host:/path', meaning: 'Securely copy a file to a remote host' },
      { cmd: 'ip route get 1.1.1.1', meaning: 'Show route to a destination IP' },
      { cmd: 'uname -r', meaning: 'Print the running kernel version' },
      { cmd: 'whoami', meaning: 'Print the current effective user' },
      { cmd: 'export VAR=value', meaning: 'Set an environment variable for the shell' },
    ],
  },
  {
    id: 'go',
    name: 'Go',
    color: '#2dd4bf',
    commands: [
      { cmd: 'go run .', meaning: 'Compile and run the current package' },
      { cmd: 'go build', meaning: 'Compile the package without running' },
      { cmd: 'go test ./...', meaning: 'Run tests in all packages' },
      { cmd: 'go mod tidy', meaning: 'Add missing and remove unused modules' },
      { cmd: 'go mod init <module>', meaning: 'Initialize a new go.mod' },
      { cmd: 'go get <pkg>', meaning: 'Add a dependency to the module' },
      { cmd: 'go fmt ./...', meaning: 'Format all Go source files' },
      { cmd: 'go vet ./...', meaning: 'Run static correctness checks' },
      { cmd: 'go install <pkg>', meaning: 'Build and install a binary to GOPATH/bin' },
      { cmd: 'go clean -modcache', meaning: 'Remove downloaded module cache' },
    ],
  },
  {
    id: 'php',
    name: 'PHP',
    color: '#fb7185',
    commands: [
      { cmd: 'php -S localhost:8000', meaning: 'Start built-in development server' },
      { cmd: 'composer install', meaning: 'Install project dependencies from composer.lock' },
      { cmd: 'composer require <pkg>', meaning: 'Add a package to composer.json' },
      { cmd: 'php artisan serve', meaning: 'Run Laravel development server' },
      { cmd: 'php -m', meaning: 'List compiled/loaded PHP modules' },
      { cmd: 'php -i', meaning: 'Print full PHP configuration (phpinfo)' },
      { cmd: 'phpunit', meaning: 'Run PHPUnit test suite' },
    ],
  },
  {
    id: 'psql',
    name: 'PostgreSQL',
    color: '#34d399',
    commands: [
      { cmd: 'psql -U <user> -d <db>', meaning: 'Connect to a database interactively' },
      { cmd: '\\l', meaning: 'List all databases (inside psql)' },
      { cmd: '\\dt', meaning: 'List tables in the current database' },
      { cmd: '\\du', meaning: 'List database roles/users' },
      { cmd: '\\q', meaning: 'Quit the psql session' },
      { cmd: 'pg_dump -U <user> <db> > out.sql', meaning: 'Export a database to SQL' },
      { cmd: 'pg_restore -d <db> dump.tar', meaning: 'Restore a database from an archive' },
    ],
  },
  {
    id: 'git',
    name: 'Git',
    color: '#fbbf24',
    commands: [
      { cmd: 'git status', meaning: 'Show working tree status' },
      { cmd: 'git add -A', meaning: 'Stage all changes' },
      { cmd: 'git commit -m "msg"', meaning: 'Create a commit with a message' },
      { cmd: 'git push origin main', meaning: 'Push commits to the remote main branch' },
      { cmd: 'git pull --rebase', meaning: 'Fetch and rebase local commits on top' },
      { cmd: 'git log --oneline -20', meaning: 'Show recent commit history condensed' },
      { cmd: 'git branch -a', meaning: 'List all local and remote branches' },
      { cmd: 'git checkout -b <branch>', meaning: 'Create and switch to a new branch' },
    ],
  },
]

export function searchCommands(query) {
  if (!query) return commandGroups
  const q = query.toLowerCase()
  return commandGroups
    .map((g) => ({
      ...g,
      commands: g.commands.filter(
        (c) => c.cmd.toLowerCase().includes(q) || c.meaning.toLowerCase().includes(q) || g.name.toLowerCase().includes(q),
      ),
    }))
    .filter((g) => g.commands.length > 0)
}
