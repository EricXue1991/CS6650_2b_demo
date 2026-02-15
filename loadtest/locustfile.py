from locust import HttpUser, task, between
import random

class ProductUser(HttpUser):
    wait_time = between(0.01, 0.1)

    def on_start(self):
        # preload some products so GET hit has data
        for i in range(50):
            pid = 1000 + i
            self.client.post(
                f"/products/{pid}/details",
                json={"sku":"X","manufacturer":"Y","category_id":1,"weight":1,"some_other_id":1},
                name="POST /products/:id/details"
            )

    @task(8)
    def get_existing(self):
        pid = 1000 + random.randint(0, 49)
        self.client.get(f"/products/{pid}", name="GET /products/:id (hit)")

    @task(2)
    def post_update(self):
        pid = 2000 + random.randint(0, 49)
        self.client.post(
            f"/products/{pid}/details",
            json={"sku":"X","manufacturer":"Y","category_id":1,"weight":1,"some_other_id":1},
            name="POST /products/:id/details"
        )
