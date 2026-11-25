import { Component, inject, signal } from '@angular/core';
import {
  FormBuilder,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { RouterLink } from '@angular/router';
import { finalize } from 'rxjs';

import { Auth } from '../../core/services/auth';

type Stage = 'form' | 'done';

@Component({
  selector: 'app-change-password',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './change-password.html',
  styleUrl: './change-password.css',
})
export class ChangePassword {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(Auth);

  readonly stage = signal<Stage>('form');
  readonly loading = signal(false);
  readonly errorMessage = signal('');

  readonly form = this.fb.nonNullable.group({
    email: [''],
    currentPassword: ['', Validators.required],
    newPassword: ['', [Validators.required, Validators.minLength(8)]],
    repeatedPassword: ['', Validators.required],
  });

  constructor() {
    const email = history.state?.email;

    if (typeof email === 'string' && email) {
      this.form.patchValue({ email });
    }
  }

  get passwordsDoNotMatch(): boolean {
    const newPassword = this.form.controls.newPassword.value;
    const repeatedPassword = this.form.controls.repeatedPassword.value;

    return repeatedPassword.length > 0 && newPassword !== repeatedPassword;
  }

  submit(): void {
    this.errorMessage.set('');

    if (this.form.invalid || this.passwordsDoNotMatch) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    this.auth.changePassword(this.form.getRawValue())
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: () => this.stage.set('done'),
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  private showError(error: HttpErrorResponse): void {
    const detail = (error.error as { detail?: string } | null)?.detail;
    this.errorMessage.set(detail || 'Request failed. Please try again.');
  }
}
