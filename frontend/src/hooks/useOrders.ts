"use client";

import useFetch from "./useFetch";
import { useMutation } from "./useMutation";
import { orderService } from "@/services/orders";

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
