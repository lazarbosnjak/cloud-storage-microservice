
import { Component, inject, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { Router, RouterLink } from '@angular/router';
import {
  FormField,
  email,
  form,
  maxLength,
  pattern,
  required,
  submit,
} from '@angular/forms/signals';
import { firstValueFrom } from 'rxjs';

import { Auth } from '../../core/services/auth';
import { decodeJwtPayload, isTokenExpired } from '../../core/utils/jwt';

type Stage = 'credentials' | 'otp' | 'done';

@Component({
  selector: 'app-sign-in',
  imports: [FormField, RouterLink],
  templateUrl: './sign-in.html',
  styleUrl: './sign-in.css',
})
export class SignIn {
  private readonly auth = inject(Auth);
  private readonly router = inject(Router);

  readonly stage = signal<Stage>('credentials');
  readonly loading = signal(false);
  readonly errorMessage = signal('');
  readonly expiresInSeconds = signal<number | null>(null);

  private otpChallenge = '';

  readonly credentialsModel = signal({ email: '', password: '' });
  readonly credentialsForm = form(this.credentialsModel, (f) => {
    required(f.email);
    email(f.email);
    required(f.password);
  });

  readonly otpModel = signal({ code: '' });
  readonly otpForm = form(this.otpModel, (f) => {
    required(f.code);
    pattern(f.code, /^\d{6}$/);
    maxLength(f.code, 6);
  });

  async submitCredentials(event: Event): Promise<void> {
    event.preventDefault();
    this.errorMessage.set('');

    this.loading.set(true);

    await submit(this.credentialsForm, async () => {
      try {
        const response = await firstValueFrom(
          this.auth.signIn(this.credentialsModel()),
        );

        if (response.mustChangePassword) {
          this.router.navigate(['/change-password'], {
            state: { email: this.credentialsModel().email },
          });
        } else if (response.otpRequired) {
          this.otpChallenge = response.otpChallenge;
          this.stage.set('otp');
        } else {
          this.completeSignIn(response.accessToken, response.expiresInSeconds);
        }
      } catch (error) {
        if (error instanceof HttpErrorResponse) {
          this.showError(error);
        } else {
          throw error;
        }
      } finally {
        this.loading.set(false);
      }

      return [];
    });
  }

  async submitOtp(event: Event): Promise<void> {
    event.preventDefault();
    this.errorMessage.set('');

    this.loading.set(true);

    await submit(this.otpForm, async () => {
      try {
        const response = await firstValueFrom(
          this.auth.verifyOtp({
            otpChallenge: this.otpChallenge,
            code: this.otpModel().code,
          }),
        );

        this.completeSignIn(response.accessToken, response.expiresInSeconds);
      } catch (error) {
        if (error instanceof HttpErrorResponse) {
          this.showError(error);
        } else {
          throw error;
        }
      } finally {
        this.loading.set(false);
      }

      return [];
    });
  }

  backToCredentials(): void {
    this.otpChallenge = '';
    this.otpForm().reset();
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
