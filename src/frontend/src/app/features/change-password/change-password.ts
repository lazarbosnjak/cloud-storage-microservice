import { Component, computed, inject, signal } from '@angular/core';
import {
  FormField,
  form,
  minLength,
  required,
  submit,
} from '@angular/forms/signals';
import { HttpErrorResponse } from '@angular/common/http';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { Auth } from '../../core/services/auth';

type Stage = 'form' | 'done';

@Component({
  selector: 'app-change-password',
  imports: [FormField, RouterLink],
  templateUrl: './change-password.html',
  styleUrl: './change-password.css',
})
export class ChangePassword {
  private readonly auth = inject(Auth);

  readonly stage = signal<Stage>('form');
  readonly loading = signal(false);
  readonly errorMessage = signal('');

  readonly model = signal({
    email: '',
    currentPassword: '',
    newPassword: '',
    repeatedPassword: '',
  });
  readonly form = form(this.model, (f) => {
    required(f.currentPassword);
    required(f.newPassword);
    minLength(f.newPassword, 8);
    required(f.repeatedPassword);
  });

  readonly passwordsDoNotMatch = computed(() => {
    const newPassword = this.form.newPassword().value();
    const repeatedPassword = this.form.repeatedPassword().value();

    return repeatedPassword.length > 0 && newPassword !== repeatedPassword;
  });

  constructor() {
    const email = history.state?.email;

    if (typeof email === 'string' && email) {
      this.model.update((m) => ({ ...m, email }));
    }
  }

  async submit(event: Event): Promise<void> {
    event.preventDefault();
    this.errorMessage.set('');

    if (this.passwordsDoNotMatch()) {
      this.form.repeatedPassword().markAsTouched();
      return;
    }

    this.loading.set(true);

    await submit(this.form, async () => {
      try {
        await firstValueFrom(
          this.auth.changePassword({
            email: this.model().email,
            currentPassword: this.model().currentPassword,
            newPassword: this.model().newPassword,
          }),
        );
        this.stage.set('done');
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

  private showError(error: HttpErrorResponse): void {
    const detail = (error.error as { detail?: string } | null)?.detail;
    this.errorMessage.set(detail || 'Request failed. Please try again.');
  }
}
