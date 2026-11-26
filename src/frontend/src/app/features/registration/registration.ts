import { Component, computed, inject, signal } from '@angular/core';
import {
  FormField,
  email,
  form,
  minLength,
  required,
  submit,
} from '@angular/forms/signals';
import { HttpErrorResponse } from '@angular/common/http';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { Auth } from '../../core/services/auth';

@Component({
  selector: 'app-registration',
  imports: [FormField, RouterLink],
  templateUrl: './registration.html',
  styleUrl: './registration.css',
})
export class Registration {
  private readonly auth = inject(Auth);

  readonly loading = signal(false);
  readonly successMessage = signal('');
  readonly errorMessage = signal('');

  readonly model = signal({
    firstName: '',
    lastName: '',
    email: '',
    password: '',
    repeatedPassword: '',
  });
  readonly form = form(this.model, (f) => {
    required(f.firstName);
    required(f.lastName);
    required(f.email);
    email(f.email);
    required(f.password);
    minLength(f.password, 8);
    required(f.repeatedPassword);
  });

  readonly passwordsDoNotMatch = computed(() => {
    const password = this.form.password().value();
    const repeatedPassword = this.form.repeatedPassword().value();

    return repeatedPassword.length > 0 && password !== repeatedPassword;
  });

  async submit(event: Event): Promise<void> {
    event.preventDefault();
    this.successMessage.set('');
    this.errorMessage.set('');

    if (this.passwordsDoNotMatch()) {
      this.form.repeatedPassword().markAsTouched();
      return;
    }

    this.loading.set(true);

    await submit(this.form, async () => {
      try {
        await firstValueFrom(this.auth.register(this.model()));
        this.successMessage.set(
          'Registration successful! Check your e-mail for the verification link.',
        );
        this.form().reset();
      } catch (error) {
        if (error instanceof HttpErrorResponse) {
          if (error.status === 409) {
            this.errorMessage.set('An account with this email already exists.');
          } else {
            this.errorMessage.set('Registration failed. Please try again.');
          }
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
