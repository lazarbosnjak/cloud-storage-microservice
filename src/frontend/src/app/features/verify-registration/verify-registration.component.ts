import { Component, inject, input, OnInit, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { Auth } from '../../core/services/auth';

@Component({
  selector: 'app-verify-registration',
  templateUrl: './verify-registration.component.html',
})
export class VerifyRegistration implements OnInit {
  private authService = inject(Auth);

  id = input.required<string>();

  loading = signal(true);
  verified = signal(false);
  errorMessage = signal('');

  ngOnInit(): void {
    this.authService.verifyRegistration(this.id()).subscribe({
      next: () => {
        this.verified.set(true);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMessage.set(
          err.error?.detail ?? 'Verification failed. The link may be invalid or expired.',
        );
        this.loading.set(false);
      },
    });
  }
}
