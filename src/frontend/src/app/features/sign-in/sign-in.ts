
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
import { decodeJwtPayload, isTokenExpired } from '../../core/utils/jwt';

type Stage = 'credentials' | 'otp' | 'done';

@Component({
  selector: 'app-sign-in',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './sign-in.html',
  styleUrl: './sign-in.css',
})
export class SignIn {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(Auth);

  readonly stage = signal<Stage>('credentials');
  readonly loading = signal(false);
  readonly errorMessage = signal('');
  readonly expiresInSeconds = signal<number | null>(null);

  private otpChallenge = '';

  readonly credentialsForm = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', Validators.required],
  });

  readonly otpForm = this.fb.nonNullable.group({
    code: ['', [Validators.required, Validators.pattern(/^\d{6}$/)]],
  });

  submitCredentials(): void {
    this.errorMessage.set('');

    if (this.credentialsForm.invalid) {
      this.credentialsForm.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    this.auth.signIn(this.credentialsForm.getRawValue())
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (response) => {
          if (response.otpRequired) {
            this.otpChallenge = response.otpChallenge;
            this.stage.set('otp');
          } else {
            this.completeSignIn(response.accessToken, response.expiresInSeconds);
          }
        },
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  submitOtp(): void {
    this.errorMessage.set('');

    if (this.otpForm.invalid) {
      this.otpForm.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    this.auth.verifyOtp({
      otpChallenge: this.otpChallenge,
      code: this.otpForm.controls.code.value,
    })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (response) =>
          this.completeSignIn(response.accessToken, response.expiresInSeconds),
        error: (error: HttpErrorResponse) => this.showError(error),
      });
  }

  backToCredentials(): void {
    this.otpChallenge = '';
    this.otpForm.reset();
    this.errorMessage.set('');
    this.stage.set('credentials');
  }

  private completeSignIn(accessToken: string, expiresInSeconds: number): void {
    const payload = decodeJwtPayload(accessToken);

    if (!accessToken || !payload || isTokenExpired(accessToken)) {
      this.errorMessage.set('Invalid session token. Please try again.');
      this.backToCredentials();
      return;
    }

    sessionStorage.setItem('access_token', accessToken);
    this.expiresInSeconds.set(expiresInSeconds);
    this.stage.set('done');
  }

  private showError(error: HttpErrorResponse): void {
    const detail = (error.error as { detail?: string } | null)?.detail;
    this.errorMessage.set(detail || 'Request failed. Please try again.');
  }
}
