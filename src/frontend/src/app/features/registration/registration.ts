import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { finalize } from 'rxjs';

import { Auth } from '../../core/services/auth';

@Component({
  selector: 'app-registration',
  imports: [ReactiveFormsModule],
  templateUrl: './registration.html',
  styleUrl: './registration.css',
})
export class Registration {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(Auth);

  readonly loading = signal(false);
  readonly successMessage = signal('');
  readonly errorMessage = signal('');

  readonly form = this.fb.nonNullable.group({
    firstName: ['', Validators.required],
    lastName: ['', Validators.required],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
    repeatedPassword: ['', Validators.required],
  });

  get passwordsDoNotMatch(): boolean {
    const password = this.form.controls.password.value;
    const repeatedPassword = this.form.controls.repeatedPassword.value;

    return repeatedPassword.length > 0 && password !== repeatedPassword;
  }

  submit(): void {
    this.successMessage.set('');
    this.errorMessage.set('');

    if (this.form.invalid || this.passwordsDoNotMatch) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    this.auth
      .register(this.form.getRawValue())
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: () => {
          this.successMessage.set(
            'Registration successful! Check your e-mail for the verification link.',
          );
          this.form.reset();
        },
        error: (error: HttpErrorResponse) => {
          if (error.status === 409) {
            this.errorMessage.set('An account with this email already exists.');
          } else {
            this.errorMessage.set('Registration failed. Please try again.');
          }
        },
      });
  }
}
