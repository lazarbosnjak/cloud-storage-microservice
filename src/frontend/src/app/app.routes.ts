import { Routes } from '@angular/router';
import { VerifyRegistration } from './features/verify-registration/verify-registration.component';

export const routes: Routes = [
  {
    path: 'sign-in',
    loadComponent: () =>
      import('./features/sign-in/sign-in').then((m) => m.SignIn),
  },
  {
    path: 'change-password',
    loadComponent: () =>
      import('./features/change-password/change-password').then(
        (m) => m.ChangePassword,
      ),
  },
  {
    path: 'forgot-password',
    loadComponent: () =>
      import('./features/forgot-password/forgot-password').then(
        (m) => m.ForgotPassword,
      ),
  },
  {
    path: 'password/reset',
    children: [
      { path: '', redirectTo: '/forgot-password', pathMatch: 'full' },
      {
        path: ':id',
        loadComponent: () =>
          import('./features/reset-password/reset-password').then(
            (m) => m.ResetPassword,
          ),
      },
    ],
  },
  {
    path: 'register',
    loadComponent: () => import('./features/registration/registration').then((m) => m.Registration),
  },
  {
    path: 'register/verify',
    children: [
      { path: '', redirectTo: '/', pathMatch: 'full' },
      { path: ':id', component: VerifyRegistration },
    ],
  },
  {
    path: '',
    redirectTo: 'sign-in',
    pathMatch: 'full',
  },
  {
  path: 'files',
  loadComponent: () =>
    import('./features/file-browser/file-browser').then(
      (m) => m.FileBrowser,
    ),
  },
];
