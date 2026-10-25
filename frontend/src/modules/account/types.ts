/**
 * User account & address models matching backend user handlers.
 */

export interface Address {
  id: string;
  userId: string;
  label?: string;
  type?: string;
  fullName: string;
  name?: string;
  phone?: string;
  line1: string;
  street?: string;
  line2?: string;
  city: string;
  state?: string;
  postalCode?: string;
  country: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAddressPayload {
  label?: string;
  fullName: string;
  phone?: string;
  line1: string;
  line2?: string;
  city: string;
  state?: string;
  postalCode?: string;
  country: string;
  isDefault?: boolean;
}

export type UpdateAddressPayload = Partial<CreateAddressPayload>;
