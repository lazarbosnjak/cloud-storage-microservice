import { Component, inject, signal } from '@angular/core';
import { FormField, email, form, required, submit } from '@angular/forms/signals';
import { HttpErrorResponse } from '@angular/common/http';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { Auth } from '../../core/services/auth';

type Stage = 'form' | 'done';

@Component({
  selector: 'app-forgot-password',
  imports: [FormField, RouterLink],
  templateUrl: './forgot-password.html',
  styleUrl: './forgot-password.css',
})
export class ForgotPassword {
  private readonly auth = inject(Auth);

  readonly stage = signal<Stage>('form');
  readonly loading = signal(false);
  readonly errorMessage = signal('');

  readonly model = signal({
    email: '',
  });
  readonly form = form(this.model, (f) => {
    required(f.email);
    email(f.email);
  });

  async submit(event: Event): Promise<void> {
    event.preventDefault();
    this.errorMessage.set('');

    this.loading.set(true);

    await submit(this.form, async () => {
      try {
        await firstValueFrom(
          this.auth.forgotPassword({
            email: this.model().email,
          }),
        );
        this.stage.set('done');
      } catch (error) {
        if (error instanceof HttpErrorResponse) {
          this.errorMessage.set('Request failed. Please try again.');
        } else {
          throw error;
        }
      } finally {
        this.loading.set(false);
      }

      return [];
    });
  }
}
