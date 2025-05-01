import pytest

fmt = lambda dep: dep.rsplit(':', 1)[0].replace('.', '_').replace(':', '_').replace('-', '_')

def test_empty_string():
    assert fmt('') == ''

def test_standard_dependency():
    assert fmt('com.google.guava:guava:30.1-jre') == 'com_google_guava_guava'
    assert fmt('org.apache.commons:commons-lang3:3.12.0') == 'org_apache_commons_commons_lang3'
    assert fmt('org.slf4j:slf4j-api:1.7.30') == 'org_slf4j_slf4j_api'

def test_group_artifact_only():
    assert fmt('org.apache.commons:commons-lang') == 'org_apache_commons'

def test_sing_segment_no_colon():
    assert fmt('com.google.guava') == 'com_google_guava'
    assert fmt('org.apache.commons') == 'org_apache_commons'