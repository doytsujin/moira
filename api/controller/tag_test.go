package controller

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/moira-alert/moira-alert"
	"github.com/moira-alert/moira-alert/api"
	"github.com/moira-alert/moira-alert/api/dto"
	"github.com/moira-alert/moira-alert/mock/moira-alert"
	"github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetAllTags(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	database := mock_moira_alert.NewMockDatabase(mockCtrl)

	Convey("Success", t, func() {
		database.EXPECT().GetTagNames().Return([]string{"tag21", "tag22", "tag1"}, nil)
		data, err := GetAllTags(database)
		So(err, ShouldBeNil)
		So(data, ShouldResemble, &dto.TagsData{TagNames: []string{"tag21", "tag22", "tag1"}})
	})

	Convey("Error", t, func() {
		expected := fmt.Errorf("Nooooooooooooooooooooo!")
		database.EXPECT().GetTagNames().Return(nil, expected)
		data, err := GetAllTags(database)
		So(err, ShouldResemble, api.ErrorInternalServer(expected))
		So(data, ShouldBeNil)
	})
}

func TestDeleteTag(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	database := mock_moira_alert.NewMockDatabase(mockCtrl)
	tag := "MyTag"

	Convey("Test no trigger ids by tag", t, func() {
		database.EXPECT().GetTagTriggerIds(tag).Return(nil, nil)
		database.EXPECT().DeleteTag(tag).Return(nil)
		resp, err := DeleteTag(database, tag)
		So(err, ShouldBeNil)
		So(resp, ShouldResemble, &dto.MessageResponse{Message: "tag deleted"})
	})

	Convey("Test has trigger ids by tag", t, func() {
		database.EXPECT().GetTagTriggerIds(tag).Return([]string{"123"}, nil)
		resp, err := DeleteTag(database, tag)
		So(err, ShouldResemble, api.ErrorInvalidRequest(fmt.Errorf("This tag is assigned to %v triggers. Remove tag from triggers first", 1)))
		So(resp, ShouldBeNil)
	})

	Convey("GetTagTriggerIds error", t, func() {
		expected := fmt.Errorf("Can not read trigger ids")
		database.EXPECT().GetTagTriggerIds(tag).Return(nil, expected)
		resp, err := DeleteTag(database, tag)
		So(err, ShouldResemble, api.ErrorInternalServer(expected))
		So(resp, ShouldBeNil)
	})

	Convey("Error delete tag", t, func() {
		expected := fmt.Errorf("Can not delete tag")
		database.EXPECT().GetTagTriggerIds(tag).Return(nil, nil)
		database.EXPECT().DeleteTag(tag).Return(expected)
		resp, err := DeleteTag(database, tag)
		So(err, ShouldResemble, api.ErrorInternalServer(expected))
		So(resp, ShouldBeNil)
	})
}

func TestGetAllTagsAndSubscriptions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	database := mock_moira_alert.NewMockDatabase(mockCtrl)
	logger, _ := logging.GetLogger("Test")

	Convey("Success get tag stats", t, func() {
		tags := []string{"tag21", "tag22", "tag1"}
		database.EXPECT().GetTagNames().Return(tags, nil)
		database.EXPECT().GetTagsSubscriptions([]string{"tag21"}).Return([]moira.SubscriptionData{{Tags: []string{"tag21"}}}, nil)
		database.EXPECT().GetTagTriggerIds("tag21").Return([]string{"trigger21"}, nil)
		database.EXPECT().GetTagsSubscriptions([]string{"tag22"}).Return(make([]moira.SubscriptionData, 0), nil)
		database.EXPECT().GetTagTriggerIds("tag22").Return([]string{"trigger22"}, nil)
		database.EXPECT().GetTagsSubscriptions([]string{"tag1"}).Return([]moira.SubscriptionData{{Tags: []string{"tag1", "tag2"}}}, nil)
		database.EXPECT().GetTagTriggerIds("tag1").Return(make([]string, 0), nil)
		stat, err := GetAllTagsAndSubscriptions(database, logger)
		So(err, ShouldBeNil)
		expected := &dto.TagsStatistics{
			List: []dto.TagStatistics{
				{TagName: "tag21", Triggers: []string{"trigger21"}, Subscriptions: []moira.SubscriptionData{{Tags: []string{"tag21"}}}},
				{TagName: "tag22", Triggers: []string{"trigger22"}, Subscriptions: make([]moira.SubscriptionData, 0)},
				{TagName: "tag1", Triggers: make([]string, 0), Subscriptions: []moira.SubscriptionData{{Tags: []string{"tag1", "tag2"}}}},
			},
		}
		So(stat, ShouldAlmostEqual(), expected)
	})

	Convey("Errors", t, func() {
		Convey("GetTagNames", func() {
			expected := fmt.Errorf("Can not get tag names")
			tags := []string{"tag21", "tag22", "tag1"}
			database.EXPECT().GetTagNames().Return(tags, expected)
			stat, err := GetAllTagsAndSubscriptions(database, logger)
			So(err, ShouldResemble, api.ErrorInternalServer(expected))
			So(stat, ShouldBeNil)
		})
	})
}