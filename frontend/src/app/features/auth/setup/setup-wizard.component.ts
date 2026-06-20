import { Component, computed, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormControl,
  FormGroup,
  ReactiveFormsModule,
  ValidatorFn,
  Validators,
} from '@angular/forms';
import { Router } from '@angular/router';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { TranslateModule } from '@ngx-translate/core';
import { NgClass } from '@angular/common';
import { AuthService } from '../../../core/auth/auth.service';
import { KeiIconComponent } from '../../../shared/components/icon/kei-icon.component';

const passwordsMatch: ValidatorFn = (group: AbstractControl) => {
  const pw  = group.get('password')?.value;
  const cpw = group.get('confirmPassword')?.value;
  return pw === cpw ? null : { passwordMismatch: true };
};

function generateTimeSlots(): string[] {
  const slots: string[] = [];
  for (let h = 0; h < 24; h++) {
    for (let m of [0, 30]) {
      slots.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
    }
  }
  return slots;
}

function passwordStrength(pw: string): 0 | 1 | 2 | 3 | 4 {
  if (!pw || pw.length < 8) return 0;
  let s = 1;
  if (/[A-Z]/.test(pw)) s++;
  if (/[0-9]/.test(pw)) s++;
  if (/[^A-Za-z0-9]/.test(pw)) s++;
  return s as 0 | 1 | 2 | 3 | 4;
}

const STRENGTH_LABELS: Record<number, string> = {
  0: '', 1: 'SETUP.STRENGTH_WEAK', 2: 'SETUP.STRENGTH_FAIR',
  3: 'SETUP.STRENGTH_GOOD', 4: 'SETUP.STRENGTH_STRONG',
};
const STRENGTH_CLASSES: Record<number, string> = {
  0: '', 1: 'strength--weak', 2: 'strength--fair',
  3: 'strength--good',       4: 'strength--strong',
};

@Component({
  selector: 'app-setup-wizard',
  standalone: true,
  imports: [
    ReactiveFormsModule, NgClass,
    MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule,
    TranslateModule, KeiIconComponent,
  ],
  templateUrl: './setup-wizard.component.html',
  styleUrl:    './setup-wizard.component.scss',
})
export class SetupWizardComponent {
  private readonly auth   = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly totalSteps = 3;
  protected step    = signal(1);
  protected loading = signal(false);
  protected error   = signal('');

  // ── Step 1: Organisation ───────────────────────────────────────────────────
  protected orgForm = new FormGroup({
    orgName:      new FormControl('', Validators.required),
    timezone:     new FormControl('UTC', Validators.required),
    workingStart: new FormControl('09:00', Validators.required),
    workingEnd:   new FormControl('17:00', Validators.required),
  });

  // ── Step 2: Admin account ──────────────────────────────────────────────────
  protected adminForm = new FormGroup(
    {
      name:            new FormControl('', Validators.required),
      email:           new FormControl('', [Validators.required, Validators.email]),
      password:        new FormControl('', [Validators.required, Validators.minLength(8)]),
      confirmPassword: new FormControl('', Validators.required),
    },
    { validators: passwordsMatch }
  );

  protected readonly timezones  = Intl.supportedValuesOf('timeZone');
  protected readonly timeSlots  = generateTimeSlots();

  protected readonly pwStrength = computed(() =>
    passwordStrength(this.adminForm.controls.password.value ?? '')
  );
  protected readonly pwStrengthLabel = computed(
    () => STRENGTH_LABELS[this.pwStrength()]
  );
  protected readonly pwStrengthClass = computed(
    () => STRENGTH_CLASSES[this.pwStrength()]
  );

  // ── Navigation ─────────────────────────────────────────────────────────────
  protected next(): void {
    if (this.step() === 1) {
      this.orgForm.markAllAsTouched();
      if (this.orgForm.valid) this.step.set(2);
    } else if (this.step() === 2) {
      this.adminForm.markAllAsTouched();
      if (this.adminForm.valid) this.doSetup();
    }
  }

  protected back(): void {
    if (this.step() > 1) this.step.update(s => s - 1);
  }

  protected goToLogin(): void {
    this.router.navigate(['/login']);
  }

  // ── Setup submission ────────────────────────────────────────────────────────
  private doSetup(): void {
    this.loading.set(true);
    this.error.set('');

    const org   = this.orgForm.getRawValue();
    const admin = this.adminForm.getRawValue();

    this.auth.setup({
      org_name:   org.orgName!,
      org_slug:   org.orgName!.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, ''),
      admin_name: admin.name!,
      email:      admin.email!,
      password:   admin.password!,
      timezone:   org.timezone!,
    }).subscribe({
      next: () => this.step.set(3),
      error: err => {
        const code = err.error?.code;
        this.error.set(code === 'SETUP_DONE' ? 'ERRORS.SETUP_DONE' : 'ERRORS.GENERIC');
        this.loading.set(false);
      },
    });
  }
}
