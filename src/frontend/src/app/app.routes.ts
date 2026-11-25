import { Routes } from '@angular/router';
import { VerifyRegistration } from './features/verify-registration/verify-registration.component';

export const routes: Routes = [
  {
    path: 'sign-in',
    loadComponent: () =>
      import('./features/sign-in/sign-in').then((m) => m.SignIn),
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
];
