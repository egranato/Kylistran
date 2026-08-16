import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { AccessService } from '../../core/auth/access.service';

@Component({
  selector: 'app-gate',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './gate.component.html',
  styleUrl: './gate.component.scss',
})
export class GateComponent {
  private readonly access = inject(AccessService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  readonly password = signal('');
  readonly error = signal(false);
  readonly checking = signal(false);

  async submit(): Promise<void> {
    this.checking.set(true);
    const ok = await this.access.unlock(this.password());
    this.checking.set(false);
    if (!ok) {
      this.error.set(true);
      return;
    }
    this.error.set(false);
    const redirect = this.route.snapshot.queryParamMap.get('redirect') ?? '/library';
    this.router.navigateByUrl(redirect);
  }
}
