#!/usr/bin/env python3
"""
Load testing client for KV store
Usage: python loadtest.py --clients 10 --operations 1000
"""

import socket
import time
import argparse
import statistics
from concurrent.futures import ThreadPoolExecutor, as_completed
from collections import defaultdict
import random


class KVClient:
    """Single connection to KV store"""

    def __init__(self, host="localhost", port=8080):
        self.host = host
        self.port = port
        self.sock = None

    def connect(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.host, self.port))

    def close(self):
        if self.sock:
            self.sock.close()

    def send_command(self, command):
        """Send command and return response with timing"""
        start = time.perf_counter()

        # Send command
        self.sock.sendall(f"{command}\n".encode())

        # Read response (up to newline)
        response = b""
        while True:
            chunk = self.sock.recv(1024)
            response += chunk
            if b"\n" in chunk:
                break

        end = time.perf_counter()
        latency_ms = (end - start) * 1000

        return response.decode().strip(), latency_ms

    def set(self, key, value):
        response, latency = self.send_command(f"SET {key} {value}")
        return response, latency

    def get(self, key):
        response, latency = self.send_command(f"GET {key}")
        return response, latency

    def delete(self, key):
        response, latency = self.send_command(f"DELETE {key}")
        return response, latency

    def exists(self, key):
        response, latency = self.send_command(f"EXISTS {key}")
        return response, latency


class LoadTester:
    """Orchestrates load testing"""

    def __init__(self, host="localhost", port=8080):
        self.host = host
        self.port = port
        self.results = {
            "set": [],
            "get": [],
            "delete": [],
            "exists": [],
        }
        self.errors = []
        self.correctness_errors = []

    def run_client(self, client_id, num_operations, read_ratio=0.8):
        """Run operations for a single client"""
        client = KVClient(self.host, self.port)
        client.connect()

        local_results = defaultdict(list)
        local_errors = []
        local_correctness = []

        # Track what we set for verification
        set_values = {}

        try:
            # PRE-POPULATE: Set all keys first so reads have data
            for i in range(num_operations):
                key = f"key_{client_id}_{i}"
                value = f"value_{client_id}_{i}"
                response, _ = client.set(key, value)  # Don't track latency for setup
                if response == "OK":
                    set_values[key] = value

            # NOW run the actual measured workload
            for i in range(num_operations):
                key = f"key_{client_id}_{i}"

                # Decide operation type
                if random.random() < read_ratio:
                    # Read operation (GET or EXISTS)
                    if random.random() < 0.8:
                        # GET
                        response, latency = client.get(key)
                        local_results["get"].append(latency)

                        # Verify correctness
                        if key in set_values:
                            expected = set_values[key]
                            if response != expected:
                                local_correctness.append(
                                    f"GET {key} returned '{response}', expected '{expected}'"
                                )
                    else:
                        # EXISTS
                        response, latency = client.exists(key)
                        local_results["exists"].append(latency)
                else:
                    # Write operation (SET or DELETE)
                    if random.random() < 0.9:
                        # SET (overwrite with new value)
                        value = f"new_value_{client_id}_{i}"
                        response, latency = client.set(key, value)
                        local_results["set"].append(latency)

                        if response != "OK":
                            local_errors.append(f"SET {key} returned '{response}'")
                        else:
                            set_values[key] = value
                    else:
                        # DELETE
                        response, latency = client.delete(key)
                        local_results["delete"].append(latency)

                        # Remove from tracking
                        set_values.pop(key, None)

        except Exception as e:
            local_errors.append(f"Client {client_id} error: {e}")
        finally:
            client.close()

        return local_results, local_errors, local_correctness

    def run(self, num_clients, operations_per_client, read_ratio=0.8):
        """Run load test with multiple clients"""
        print(f"Starting load test:")
        print(f"  Clients: {num_clients}")
        print(f"  Operations per client: {operations_per_client}")
        print(f"  Total operations: {num_clients * operations_per_client}")
        print(f"  Read ratio: {read_ratio:.0%}")
        print()

        start_time = time.time()

        # Run clients concurrently
        with ThreadPoolExecutor(max_workers=num_clients) as executor:
            futures = [
                executor.submit(self.run_client, i, operations_per_client, read_ratio)
                for i in range(num_clients)
            ]

            for future in as_completed(futures):
                local_results, local_errors, local_correctness = future.result()

                # Aggregate results
                for op_type, latencies in local_results.items():
                    self.results[op_type].extend(latencies)

                self.errors.extend(local_errors)
                self.correctness_errors.extend(local_correctness)

        end_time = time.time()
        duration = end_time - start_time

        self.report(duration)

    def calculate_percentiles(self, latencies):
        """Calculate percentile latencies"""
        if not latencies:
            return None, None, None, None

        sorted_latencies = sorted(latencies)
        return (
            statistics.mean(latencies),
            sorted_latencies[int(len(sorted_latencies) * 0.50)],
            sorted_latencies[int(len(sorted_latencies) * 0.95)],
            sorted_latencies[int(len(sorted_latencies) * 0.99)],
        )

    def report(self, duration):
        """Print test results"""
        print("=" * 60)
        print("LOAD TEST RESULTS")
        print("=" * 60)
        print()

        total_ops = sum(len(latencies) for latencies in self.results.values())
        throughput = total_ops / duration

        print(f"Duration: {duration:.2f}s")
        print(f"Total operations: {total_ops}")
        print(f"Throughput: {throughput:.2f} ops/sec")
        print()

        # Per-operation latencies
        print("Latency by operation type:")
        print("-" * 60)
        print(
            f"{'Operation':<12} {'Count':<10} {'Mean':<10} {'P50':<10} {'P95':<10} {'P99':<10}"
        )
        print("-" * 60)

        for op_type, latencies in self.results.items():
            if latencies:
                mean, p50, p95, p99 = self.calculate_percentiles(latencies)
                print(
                    f"{op_type.upper():<12} {len(latencies):<10} "
                    f"{mean:>8.2f}ms {p50:>8.2f}ms {p95:>8.2f}ms {p99:>8.2f}ms"
                )

        print()

        # Overall latency
        all_latencies = []
        for latencies in self.results.values():
            all_latencies.extend(latencies)

        if all_latencies:
            mean, p50, p95, p99 = self.calculate_percentiles(all_latencies)
            print(
                f"{'OVERALL':<12} {len(all_latencies):<10} "
                f"{mean:>8.2f}ms {p50:>8.2f}ms {p95:>8.2f}ms {p99:>8.2f}ms"
            )

        print()

        # Errors
        if self.errors:
            print(f"Errors: {len(self.errors)}")
            for error in self.errors[:10]:  # Show first 10
                print(f"  - {error}")
            if len(self.errors) > 10:
                print(f"  ... and {len(self.errors) - 10} more")
            print()

        # Correctness errors
        if self.correctness_errors:
            print(f"Correctness errors: {len(self.correctness_errors)}")
            for error in self.correctness_errors[:10]:
                print(f"  - {error}")
            if len(self.correctness_errors) > 10:
                print(f"  ... and {len(self.correctness_errors) - 10} more")
        else:
            print("All correctness checks passed!")

        print()


def main():
    parser = argparse.ArgumentParser(description="Load test KV store")
    parser.add_argument("--host", default="localhost", help="Server host")
    parser.add_argument("--port", type=int, default=8080, help="Server port")
    parser.add_argument(
        "--clients", type=int, default=10, help="Number of concurrent clients"
    )
    parser.add_argument(
        "--operations", type=int, default=1000, help="Operations per client"
    )
    parser.add_argument(
        "--read-ratio",
        type=float,
        default=0.8,
        help="Ratio of read operations (0.0-1.0)",
    )

    args = parser.parse_args()

    tester = LoadTester(args.host, args.port)
    tester.run(args.clients, args.operations, args.read_ratio)


if __name__ == "__main__":
    main()
