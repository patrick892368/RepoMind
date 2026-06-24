package com.example;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping({"/api/v1", "/internal"})
public class OrderController {
    @GetMapping({"/orders", "/purchases"})
    public String list() {
        return "ok";
    }

    @RequestMapping(path = {"/orders/{id}", "/purchases/{id}"}, method = {RequestMethod.PUT, RequestMethod.PATCH})
    public String update() {
        return "ok";
    }

    @PostMapping(path = "/orders")
    public String create() {
        return "ok";
    }

    @PatchMapping(value = "/orders/{id}/status")
    public String status() {
        return "ok";
    }
}
