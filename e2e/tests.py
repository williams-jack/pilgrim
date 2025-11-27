from enum import Enum, auto
import os
import re
import subprocess
from sys import stderr

# TODO: Abstract test logic to support multiple DBs (Postgres, MySQL, SQLite, etc.)

INITIAL_HEALTH_CHECK_DELAY_SEC = 10
HEALTH_CHECK_RETRIES = 3

def migration_name_from_path(path: str) -> str:
    base_name = os.path.basename(path)
    # TODO: Make extension configurable for different DBs
    match = re.match(r"^[0-9]+_(.+)\.sql$", base_name)
    if match:
        return match.group(1)
    else:
        raise ValueError(f"Invalid migration file name: {base_name}")

def pg_healthy_check(config: dict) -> subprocess.CompletedProcess:
    """
    Check if Postgres DB is healthy by attempting a connection.
    """
    from time import sleep
    db_name = config["dbConfig"]["dbName"]
    user = config["dbConfig"]["user"]
    password = config["dbConfig"]["password"]
    env ={ "PGPASSWORD": password }
    check_cmd = [
        "docker",
        "exec",
        "pg",
        "psql",
        "-U",
        user,
        "-d",
        db_name,
        "-c",
        "\\q"
    ]
    health_check_retry_count = 0
    health_check_delay_sec = INITIAL_HEALTH_CHECK_DELAY_SEC
    while True:
        health_check_result = subprocess.run(check_cmd,
                                             capture_output=True,
                                             text=True,
                                             env=env)
        success_condition = health_check_result.returncode == 0
        failure_condition = health_check_retry_count >= HEALTH_CHECK_RETRIES
        if success_condition or failure_condition:
            return health_check_result
        else:
            health_check_retry_count += 1
            sleep(health_check_delay_sec)
            health_check_delay_sec *= 2  # Exponential backoff

class Tests(Enum):
    CREATE_MIGRATION = auto()
    INIT_MIGRATION_TABLE = auto()
    APPLY_MIGRATION = auto()
    ROLLBACK_MIGRATION = auto()

def test_dependencies(t: Tests) -> list[Tests]:
    mapping = {
        Tests.CREATE_MIGRATION: [],
        Tests.INIT_MIGRATION_TABLE: [],
        Tests.APPLY_MIGRATION: [Tests.INIT_MIGRATION_TABLE, Tests.CREATE_MIGRATION],
        Tests.ROLLBACK_MIGRATION: [Tests.APPLY_MIGRATION],
    }
    dependencies = mapping.get(t, [])
    i = 0
    while i < len(dependencies):
        dep = dependencies[i]
        dep_deps = mapping.get(dep, [])
        for d in dep_deps:
            if d not in dependencies:
                dependencies.append(d)
        i += 1
    return dependencies


def str_to_test(s: str) -> Tests:
    mapping = {
        "init_migration_table": Tests.INIT_MIGRATION_TABLE,
        "create_migration": Tests.CREATE_MIGRATION,
        "apply_migration": Tests.APPLY_MIGRATION,
        "rollback_migration": Tests.ROLLBACK_MIGRATION,
    }
    if s in mapping:
        return mapping[s]
    else:
        raise ValueError(f"Unknown test name: {s}")

def tests_enum_to_test_fn(t: Tests, config: dict, **kwargs):
    mapping = {
        Tests.CREATE_MIGRATION: lambda: create_migration_test(config["upDir"], config["downDir"], "test_migration"),
        Tests.INIT_MIGRATION_TABLE: lambda: init_migration_table_test(config),
        Tests.APPLY_MIGRATION: lambda: apply_migration_test(config, migration_up_file_path=kwargs.get("migration_up_file_path", "")),
        Tests.ROLLBACK_MIGRATION: lambda: rollback_migration_test(config, migration_down_file_path=kwargs.get("migration_down_file_path", "")),
    }
    return mapping.get(t, None)

def create_migration_test(up_dir: str, down_dir: str, migration_name: str) -> tuple[str, str]:
    """
    Test creating a new migration.
    """
    create_cmd = ["./pilgrim", "create", "-d", migration_name]
    result = subprocess.run(create_cmd,
                            capture_output=True,
                            text=True)
    if result.returncode != 0:
        raise Exception(f"Create migration command failed: {result.stderr}")
    migration_files_up = os.listdir(up_dir)
    migration_files_down = os.listdir(down_dir)
    regex_pattern = re.compile(r"^[0-9]+_" + re.escape(migration_name) + r"\.sql$")
    assert any(regex_pattern.match(f) for f in migration_files_up), \
        "Migration up file not created."
    assert any(regex_pattern.match(f) for f in migration_files_down), \
        "Migration down file not created."
    up_file, down_file = migration_files_up[0], migration_files_down[0]
    sample_sql_content_dir = os.path.join(
            os.path.dirname(os.path.abspath(__file__)),
            "migration-scripts",
            "sql"
    )
    up_file_path = os.path.join(up_dir, up_file)
    down_file_path = os.path.join(down_dir, down_file)
    with open(up_file_path, "w") as f:
        with open(os.path.join(sample_sql_content_dir, "up.sql"), "r") as sample_f:
            f.write(sample_f.read())
    with open(down_file_path, "w") as f:
        with open(os.path.join(sample_sql_content_dir, "down.sql"), "r") as sample_f:
            f.write(sample_f.read())
    return up_file_path, down_file_path


def init_migration_table_test(config: dict) -> None:
    """
    Test initializing the migration table in the database.
    """
    init_cmd = ["./pilgrim", "init"]
    result = subprocess.run(init_cmd,
                            capture_output=True,
                            text=True)
    if result.returncode != 0:
        raise Exception(f"Init migration table command failed: {result.stderr}")
    psql_assert_table_existance(config, "pilgrim_migration_history")

# TODO: Assert changes made by migration applied successfully.
def apply_migration_test(config: dict, migration_up_file_path: str) -> None:
    apply_cmd = ["./pilgrim", "apply"]
    result = subprocess.run(apply_cmd,
                            capture_output=True,
                            text=True)
    if result.returncode != 0:
        raise Exception(f"Apply migration command failed: {result.stderr}")
    check_apply_sql_cmd = "SELECT COUNT(*) FROM pilgrim_migration_history WHERE " + \
            f"migration_name = '{migration_name_from_path(migration_up_file_path)}';"
    result = psql_exec_command(config, check_apply_sql_cmd)
    if result.returncode != 0:
        raise Exception(f"Check applied migration command failed: {result.stderr}")
    assert "1" in result.stdout, "Migration was not applied successfully."
    table_name = "test_table"
    psql_assert_table_existance(config, table_name)

def psql_exec_command(config: dict, sql_cmd: str) -> subprocess.CompletedProcess:
    env ={ "PGPASSWORD": config["dbConfig"]["password"] }
    psql_cmd = [
        "docker",
        "exec",
        "pg",
        "psql",
        "-U",
        config["dbConfig"]["user"],
        "-d",
        config["dbConfig"]["dbName"],
        "-c",
        sql_cmd
    ]
    result = subprocess.run(psql_cmd,
                            capture_output=True,
                            text=True,
                            env=env)
    return result

def psql_assert_table_existance(config: dict, table_name: str, should_exist=True) -> None:
    check_table_sql_cmd = f"SELECT to_regclass('{table_name}');"
    result = psql_exec_command(config, check_table_sql_cmd)
    if result.returncode != 0:
        raise Exception(f"Check table exists command failed: {result.stderr}")
    if should_exist:
        assert table_name in result.stdout, f"Table {table_name} was not found in db"
    else:
        assert table_name not in result.stdout, f"Table {table_name} was found in db"

def rollback_migration_test(config: dict, migration_down_file_path: str) -> None:
    rollback_cmd = ["./pilgrim", "rollback"]
    result = subprocess.run(rollback_cmd,
                            capture_output=True,
                            text=True)
    if result.returncode != 0:
        raise Exception(f"Rollback migration command failed: {result.stderr}")
    check_history_sql_cmd = "SELECT COUNT(*) FROM pilgrim_migration_history " + \
            f"WHERE migration_name = '{migration_name_from_path(migration_down_file_path)}';"
    result = psql_exec_command(config, check_history_sql_cmd)
    if result.returncode != 0:
        raise Exception(f"Check rolled back migration command failed: {result.stderr}")
    assert "0" in result.stdout, "Migration was not rolled back successfully."
    psql_assert_table_existance(config, "test_table", should_exist=False)

def setup(use_docker: bool) -> None:
    """
    Builds pilgrim executable and spins up docker containers for DBs used
    in tests.
    """
    build_cmd = ["go", "build", "-o", "pilgrim", "../cmd/pilgrim/main.go"]
    result = subprocess.run(build_cmd,
                            capture_output=True,
                            text=True)
    if result.returncode != 0:
        raise Exception(f"Build command failed: {result.stderr}")
    if use_docker:
        up_cmd = ["docker", "compose", "up", "-d", "--build"]
        result = subprocess.run(up_cmd,
                                capture_output=True,
                                text=True)
        if result.returncode != 0:
            raise Exception(f"Docker compose up command failed: {result.stderr}")

def teardown(use_docker: bool) -> None:
    """
    Tears down docker containers used in tests.
    """
    from shutil import rmtree
    os.remove("pilgrim")
    rmtree("migrations", ignore_errors=True)
    if use_docker:
        down_cmd = ["docker", "compose", "down", "-v"]
        result = subprocess.run(down_cmd,
                                capture_output=True,
                                text=True)
        if result.returncode != 0:
            raise Exception(f"Docker compose down command failed: {result.stderr}")

def try_parse_tests(tests: list[str]) -> tuple[list[Tests], list[str]]:
    parsed_tests = []
    invalid_tests = []
    for test in tests:
        try:
            parsed_tests.append(str_to_test(test))
        except ValueError:
            invalid_tests.append(test)

    return parsed_tests, invalid_tests

if __name__ == "__main__":
    import json
    import sys
    with open("pilgrim.config.json", "r") as f:
        config = json.load(f)
    if len(sys.argv) == 1:
        tests_to_run = list(Tests)
    else:
        tests_to_run, invalid_tests = try_parse_tests(sys.argv[1:])
        if len(invalid_tests) > 0:
            stderr.write(f"Invalid test names: {', '.join(invalid_tests)}\n")
            stderr.write("Valid test names are (case-insensitive):\n")
            for test in Tests:
                stderr.write(f" - {test.name.lower()}\n")
            sys.exit(1)
    tests_to_run.sort(key=lambda x: x.value) # Ensure order for apply and rollback tests
    use_docker = len(tests_to_run) > 1 or tests_to_run[0] != Tests.CREATE_MIGRATION

    try:
        setup(use_docker)
    except subprocess.CalledProcessError as e:
        stderr.write(f"Failed to set up test dbs: {e}\n")
        sys.exit(1)

    if use_docker:
        health_check_result = pg_healthy_check(config)
        if health_check_result.returncode != 0:
            stderr.write(f"Postgres health check failed: {health_check_result.stderr}\n")
            try:
                teardown(use_docker)
            except subprocess.CalledProcessError as e:
                stderr.write(f"Failed to tear down test dbs: {e}\n")
            sys.exit(1)

    failed_tests = []
    migration_up_file_path = ""
    migration_down_file_path = ""
    for test in tests_to_run:
        try:
            test_deps = test_dependencies(test)
            if any(dep in failed_tests for dep in test_deps):
                print(f"Skipping test {test.name.lower()} due to failed dependencies " + \
                        ", ".join([dep.name.lower() for dep in test_deps if dep in failed_tests]) + ".")
                failed_tests.append(test)
                continue
            test_fn = tests_enum_to_test_fn(test,
                                            config,
                                            migration_up_file_path=migration_up_file_path,
                                            migration_down_file_path=migration_down_file_path)
            if test_fn is None:
                print(f"Test {test.name.lower()} not implemented, skipping.")
                continue
            test_result = test_fn()
            if test == Tests.CREATE_MIGRATION:
                migration_up_file_path, migration_down_file_path = test_result
            print(f"Test {test.name.lower()} passed.")
        except Exception as e:
            failed_tests.append(test)
            stderr.write(f"Test {test.name.lower()} failed: {e}\n")
    try:
        teardown(use_docker)
    except subprocess.CalledProcessError as e:
        stderr.write(f"Failed to tear down test dbs: {e}\n")

    if len(failed_tests) > 0:
        sys.exit(1)
    else:
        print("All tests passed.")
        sys.exit(0)
