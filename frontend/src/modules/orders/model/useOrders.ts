"use client";

import useFetch from "@/shared/api/hooks/useFetch";
import { useMutation } from "@/shared/api/hooks/useMutation";
import { orderService } from "@/modules/orders/api/orderService";

export function useOrders() {
  const { data, loading, errorMsg, refetch } = useFetch(() => orderService.getOrders(), "");
  return { orders: data, loading, errorMsg, refetch };
}

export function useOrder(id: string) {
  const { data, loading, errorMsg, refetch } = useFetch(() => orderService.getOrderById(id), id);
  return { order: data, loading, errorMsg, refetch };
}

export function useCheckout() {
  const { mutate, loading, errorMsg } = useMutation((payload: any) => orderService.checkout(payload));
  return { checkout: mutate, loading, errorMsg };
}
