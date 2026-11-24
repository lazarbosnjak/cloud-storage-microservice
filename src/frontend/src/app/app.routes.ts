import { Routes } from '@angular/router';
import { VerifyRegistration } from './features/verify-registration/verify-registration.component';

export const routes: Routes = [
  {
    path: 'register',
    loadComponent: () => import('./features/registration/registration').then((m) => m.Registration),
  },
  {
    path: 'register/verify/:id',
    component: VerifyRegistration,
  },
  {
    path: '',
    redirectTo: 'register',
    pathMatch: 'full',
  },
];
