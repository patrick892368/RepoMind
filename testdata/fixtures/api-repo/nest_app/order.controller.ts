import { Body, Controller, Get, Post } from "@nestjs/common";

@Controller("order")
export class OrderController {
  @Post("create")
  create(@Body() body: unknown) {
    return body;
  }

  @Get("status")
  status() {
    return { ok: true };
  }
}
