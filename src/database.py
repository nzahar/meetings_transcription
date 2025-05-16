from peewee import *
import config
import datetime

db = SqliteDatabase(config.db_file)

class BaseModel(Model):
    class Meta:
        database = db

class Users(BaseModel):
    user_id = IntegerField(column_name='user_id', unique=True)
    user_name = TextField(column_name='user_name', null=False)

    class Meta:
        table_name = 'users'

class UserRequestLogs(BaseModel):
    user_id = ForeignKeyField(Users, backref='requests', on_delete='CASCADE')
    hour = DateTimeField(column_name='req_hour', null=False)
    request_count = IntegerField(column_name='request_count', default=1)

    class Meta:
        table_name = 'user_request_logs'  # Исправленное имя

#db.connect()
#db.create_tables([Users, UserRequestLogs])


def is_user_rate_limited(user_id, max_requests=100):
    current_hour = datetime.datetime.now().replace(minute=0, second=0, microsecond=0)

    log = (UserRequestLogs
           .select(UserRequestLogs.request_count)
           .where((UserRequestLogs.user_id == user_id) & (UserRequestLogs.hour == current_hour))
           .first())

    request_count = log.request_count if log else 0

    if request_count >= max_requests:
        return True
    return False


def log_request(user_id):
    current_hour = datetime.datetime.now().replace(minute=0, second=0, microsecond=0)

    log, created = UserRequestLogs.get_or_create(
        user_id=user_id, hour=current_hour,
        defaults={'request_count': 1}
    )

    if not created:
        log.request_count += 1
        log.save()


def cleanup_old_logs():
    threshold_date = datetime.datetime.now() - datetime.timedelta(days=7)
    deleted_rows = UserRequestLogs.delete().where(UserRequestLogs.hour < threshold_date).execute()

