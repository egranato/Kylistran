import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { EditorAuthService } from '../../../core/editor-auth/editor-auth.service';

@Component({
  selector: 'app-editor-login',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './editor-login.component.html',
  styleUrl: './editor-login.component.scss',
})
export class EditorLoginComponent {
  private readonly auth = inject(EditorAuthService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  readonly username = signal('');
  readonly password = signal('');
  readonly error = signal(false);
  readonly checking = signal(false);

  async submit(): Promise<void> {
    this.checking.set(true);
    const ok = await this.auth.login(this.username(), this.password());
    this.checking.set(false);
    if (!ok) {
      this.error.set(true);
      return;
    }
    this.error.set(false);
    const redirect = this.route.snapshot.queryParamMap.get('redirect') ?? '/editor';
    this.router.navigateByUrl(redirect);
  }
}
