import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

export default function globalSetup(): void {
  execFileSync('go', ['build', '-o', 'bin/e2eserver', './cmd/e2eserver'], {
    cwd: fileURLToPath(new URL('../../backend/', import.meta.url)),
    stdio: 'inherit',
  })
}
