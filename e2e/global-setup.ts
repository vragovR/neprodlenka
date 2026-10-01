import { execFileSync } from 'node:child_process'

export default function globalSetup(): void {
  execFileSync('go', ['build', '-o', 'bin/e2eserver', './cmd/e2eserver'], { stdio: 'inherit' })
}
